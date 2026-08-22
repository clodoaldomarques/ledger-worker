resource "aws_sqs_queue" "balance-sqs-queue" {
  name = "balance-sqs-queue"
}

resource "aws_sqs_queue" "balance-sqs-dlq-queue" {
  name = "balance-sqs-dlq-queue"
}

