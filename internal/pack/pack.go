package pack

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/davidahmann/relia/internal/ledger"
	"github.com/davidahmann/relia/pkg/types"
)

const PackSchema = "relia.pack.v0.1"

type ApprovalRecord struct {
	ApprovalID string `json:"approval_id"`
	Status     string `json:"status"`
	ReceiptID  string `json:"receipt_id"`
}

type Input struct {
	Receipt   ledger.StoredReceipt
	Context   types.ContextRecord
	Decision  types.DecisionRecord
	Policy    []byte
	Approvals []ApprovalRecord
	CreatedAt string
}

func BuildZip(input Input, baseURL string) ([]byte, error) {
	files, err := BuildFiles(input, baseURL)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(nil)
	if err := WriteZip(buf, files); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func BuildFiles(input Input, baseURL string) (map[string][]byte, error) {
	if len(input.Policy) == 0 {
		return nil, fmt.Errorf("policy bytes missing")
	}

	receiptJSON, err := buildReceiptJSON(input.Receipt, baseURL)
	if err != nil {
		return nil, err
	}
	contextJSON, err := json.MarshalIndent(input.Context, "", "  ")
	if err != nil {
		return nil, err
	}
	decisionJSON, err := json.MarshalIndent(input.Decision, "", "  ")
	if err != nil {
		return nil, err
	}

	files := map[string][]byte{
		"receipt.json":  append(receiptJSON, '\n'),
		"context.json":  append(contextJSON, '\n'),
		"decision.json": append(decisionJSON, '\n'),
		"policy.yaml":   append(bytes.TrimRight(input.Policy, "\n"), '\n'),
	}

	if len(input.Approvals) > 0 {
		approvalsJSON, err := json.MarshalIndent(input.Approvals, "", "  ")
		if err != nil {
			return nil, err
		}
		files["approvals.json"] = append(approvalsJSON, '\n')
	}

	summary, summaryHTML, err := BuildSummary(input, baseURL)
	if err != nil {
		return nil, err
	}
	summaryJSON, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return nil, err
	}
	files["summary.json"] = append(summaryJSON, '\n')
	files["summary.html"] = append(summaryHTML, '\n')

	fileEntries := buildFileEntries(files)

	createdAt := input.CreatedAt
	if createdAt == "" {
		createdAt = time.Now().UTC().Format(time.RFC3339)
	}

	manifest := types.PackManifest{
		Schema:     PackSchema,
		CreatedAt:  createdAt,
		ReceiptID:  input.Receipt.ReceiptID,
		ContextID:  input.Context.ContextID,
		DecisionID: input.Decision.DecisionID,
		PolicyHash: input.Receipt.PolicyHash,
		Schemas: types.PackSchemas{
			Context:  input.Context.Schema,
			Decision: input.Decision.Schema,
			Receipt:  ledger.ReceiptSchema,
		},
		Files: fileEntries,
	}
	manifest.Refs = extractReceiptRefs(input.Receipt.BodyJSON)
	manifest.InteractionRef = extractReceiptInteractionRef(input.Receipt.BodyJSON)

	if input.Receipt.ApprovalID != nil {
		manifest.ApprovalID = *input.Receipt.ApprovalID
	}

	manifestJSON, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, err
	}
	files["manifest.json"] = append(manifestJSON, '\n')

	checksums := buildChecksums(files)
	files["sha256sums.txt"] = checksums

	return files, nil
}

func extractReceiptRefs(body []byte) *types.ReceiptRefs {
	if len(body) == 0 {
		return nil
	}
	var payload struct {
		Refs *types.ReceiptRefs `json:"refs,omitempty"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil
	}
	if payload.Refs == nil {
		return nil
	}
	if payload.Refs.Context == nil && payload.Refs.Decision == nil {
		return nil
	}
	return payload.Refs
}

func extractReceiptInteractionRef(body []byte) *types.InteractionRef {
	if len(body) == 0 {
		return nil
	}
	var payload struct {
		InteractionRef *types.InteractionRef `json:"interaction_ref,omitempty"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil
	}
	return payload.InteractionRef
}

func buildReceiptJSON(receipt ledger.StoredReceipt, baseURL string) ([]byte, error) {
	var body map[string]any
	if err := json.Unmarshal(receipt.BodyJSON, &body); err != nil {
		return nil, err
	}

	sig := "base64:" + base64.StdEncoding.EncodeToString(receipt.Sig)
	body["integrity"] = map[string]any{
		"body_digest": receipt.BodyDigest,
		"signatures": []map[string]any{
			{
				"alg":    "Ed25519",
				"key_id": receipt.KeyID,
				"sig":    sig,
			},
		},
	}

	if baseURL != "" {
		body["links"] = map[string]any{
			"verify": strings.TrimRight(baseURL, "/") + "/v1/verify/" + receipt.ReceiptID,
			"pack":   strings.TrimRight(baseURL, "/") + "/v1/pack/" + receipt.ReceiptID,
		}
	}

	return json.MarshalIndent(body, "", "  ")
}

func buildFileEntries(files map[string][]byte) []types.PackFile {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)

	entries := make([]types.PackFile, 0, len(names))
	for _, name := range names {
		sum := sha256.Sum256(files[name])
		entries = append(entries, types.PackFile{
			Name:      name,
			SHA256:    "sha256:" + hex.EncodeToString(sum[:]),
			SizeBytes: int64(len(files[name])),
		})
	}
	return entries
}

func buildChecksums(files map[string][]byte) []byte {
	names := make([]string, 0, len(files))
	for name := range files {
		if name == "sha256sums.txt" {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)

	var buf bytes.Buffer
	for _, name := range names {
		sum := sha256.Sum256(files[name])
		_, _ = fmt.Fprintf(&buf, "sha256:%s  %s\n", hex.EncodeToString(sum[:]), name)
	}
	return buf.Bytes()
}

func WriteZip(w io.Writer, files map[string][]byte) error {
	writer := zip.NewWriter(w)
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		entry, err := writer.Create(name)
		if err != nil {
			_ = writer.Close()
			return err
		}
		if _, err := entry.Write(files[name]); err != nil {
			_ = writer.Close()
			return err
		}
	}

	return writer.Close()
}
