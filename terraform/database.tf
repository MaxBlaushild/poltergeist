resource "aws_security_group" "db_sg" {
  name        = "db_sg"
  description = "Allow inbound traffic"
  vpc_id      = module.vpc.vpc_id

  ingress {
    description = "PostgreSQL"
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_db_instance" "poltergeist-db" {
  identifier             = "poltergeist"
  engine                 = "postgres"
  engine_version         = "15.12"
  instance_class         = "db.t3.micro"
  allocated_storage      = 20
  db_name                = "poltergeist"
  username               = "db_user"
  password               = var.db_password
  skip_final_snapshot    = true
  db_subnet_group_name   = aws_db_subnet_group.main.name
  vpc_security_group_ids = [aws_security_group.db_sg.id]

  parameter_group_name    = "default.postgres15"
  backup_retention_period = 1
}

