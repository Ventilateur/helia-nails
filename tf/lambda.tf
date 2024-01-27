locals {
  function_name = "calendar-sync"
}

data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

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
    effect  = "Allow"
    actions = [
      "ssm:GetParameter",
      "kms:Decrypt"
    ]
    resources = [
      "*"
    ]
  }
}

data "aws_iam_policy_document" "cloudwatch_logs" {
  statement {
    effect  = "Allow"
    actions = [
      "logs:CreateLogGroup"
    ]
    resources = [
      "arn:aws:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:*"
    ]
  }

  statement {
    effect  = "Allow"
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]
    resources = [
      "arn:aws:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:log-group:/aws/lambda/${local.function_name}:*"
    ]
  }
}

resource "aws_iam_role" "calendar_sync_lambda" {
  name               = "calendar_sync_lambda"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json

  inline_policy {
    name   = "get-parameters-lambda"
    policy = data.aws_iam_policy_document.get_parameters.json
  }

  inline_policy {
    name   = "cloudwatch-logs-lambda"
    policy = data.aws_iam_policy_document.cloudwatch_logs.json
  }
}

resource "aws_lambda_function" "calendar_sync" {
  package_type  = "Image"
  image_uri     = "${aws_ecr_repository.calendar-sync.repository_url}:${var.image_version}"
  function_name = local.function_name
  role          = aws_iam_role.calendar_sync_lambda.arn
  timeout       = 300
}

resource "aws_iam_role" "event_bridge_execution_role" {
  name               = "event_bridge_execution_role"
  assume_role_policy = jsonencode(
    {
      "Version" : "2012-10-17",
      "Statement" : [
        {
          "Effect" : "Allow",
          "Principal" : {
            "Service" : "scheduler.amazonaws.com"
          },
          "Action" : "sts:AssumeRole"
        }
      ]
    }
  )

  inline_policy {
    name   = "invoke-lambda"
    policy = jsonencode(
      {
        "Version" : "2012-10-17",
        "Statement" : [
          {
            "Action" : [
              "lambda:InvokeFunction"
            ],
            "Effect" : "Allow",
            "Resource" : aws_lambda_function.calendar_sync.arn
          }
        ]
      }
    )
  }
}

resource "aws_scheduler_schedule" "example" {
  name       = "calendar-sync"
  group_name = "default"

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression = "rate(3 minutes)"

  target {
    arn      = aws_lambda_function.calendar_sync.arn
    role_arn = aws_iam_role.event_bridge_execution_role.arn
    input    = jsonencode(
      {
        name : "sync"
      }
    )
  }
}
