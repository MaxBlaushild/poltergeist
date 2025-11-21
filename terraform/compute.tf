resource "aws_key_pair" "lappentoppen" {
  key_name   = "lappentoppen"
  public_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIHR4WRxoB3jd6Q/qFXBIgKqkPwo9gXyzUHctXpZgeMx0"
}

resource "aws_security_group" "allow_ssh" {
  name        = "allow_ssh"
  description = "Allow SSH inbound traffic"
  vpc_id      = module.vpc.vpc_id

  ingress {
    from_port   = 22
    to_port     = 22
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

resource "aws_instance" "ssh_box" {
  ami                    = "ami-05c13eab67c5d8861" // free cheap linux box
  instance_type          = "t2.nano"
  key_name               = aws_key_pair.lappentoppen.key_name
  subnet_id              = module.vpc.public_subnets[0]
  vpc_security_group_ids = [aws_security_group.allow_ssh.id, aws_security_group.db_sg.id, aws_security_group.redis_sg.id]
  associate_public_ip_address = true

  tags = {
    Name = "SshBox"
  }
}

