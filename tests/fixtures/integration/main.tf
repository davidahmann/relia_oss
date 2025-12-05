provider "aws" {
  region = "us-west-2"
}

resource "aws_instance" "web" {
  instance_type = "t3.micro"
}

resource "aws_nat_gateway" "main" {
  connectivity_type = "public"
}

resource "aws_lambda_function" "api" {
  function_name = "api"
  memory_size   = 128
}
