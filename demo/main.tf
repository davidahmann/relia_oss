provider "aws" {
  region = "us-east-1"
}

resource "aws_instance" "web" {
  instance_type = "t3.large"
  # t3.large in us-east-1 is typically ~$0.0832/hr -> $60/mo
}

resource "aws_instance" "db" {
  instance_type = "m5.large"
  # m5.large is typically ~$0.096/hr -> $70/mo
}
