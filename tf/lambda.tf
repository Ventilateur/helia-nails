data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

data "aws_iam_policy_document" "get_parameters" {
  statement {
    effect = "Allow"
    actions = [
      "ssm:GetParameter",
      "kms:Decrypt"
    ]
    resources = [
      "*"
    ]
  }
}

resource "aws_iam_role" "calendar_sync_lambda" {
  name               = "calendar_sync_lambda"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
  inline_policy {
    name = "get-parameters-lambda"
    policy = data.aws_iam_policy_document.get_parameters.json
  }
}

resource "aws_lambda_function" "calendar_sync" {
  package_type  = "Image"
  image_uri     = "${aws_ecr_repository.calendar-sync.repository_url}:${var.image_version}"
  function_name = "calendar-sync"
  role          = aws_iam_role.calendar_sync_lambda.arn
}
