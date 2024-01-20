resource "aws_ecr_repository" "calendar-sync" {
  name                 = "helia-nails/calendar-sync"
  image_tag_mutability = "MUTABLE"
}
