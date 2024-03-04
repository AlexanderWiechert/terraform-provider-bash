terraform {
  required_providers {
    bash = {
      source = "apparentlymart/bash"
    }
    local = {
      source  = "hashicorp/local"
      version = "~> 2.4.1"
    }
  }
}

resource "local_file" "example" {
  filename = "${path.module}/example.sh"
  content  = provider::bash::script(file("${path.module}/example.sh.tmpl"), {
    greeting = "Hello"
    names    = tolist(["Medhi", "Aurynn", "Kat", "Ariel"])
    num      = 3
    ids = tomap({
      a = "i-123"
      b = "i-456"
      c = "i-789"
    })
  })
}

output "output_filename" {
  value = local_file.example.filename
}
