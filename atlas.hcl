env "ent" {
  src = "./internal/shared/infrastructure/ent/schema"
  dev = "docker://postgres/16/dev"

  migration {
    dir = "file://migrations"
  }

  format {
    migrate {
      diff = "{{ sql . \"  \" }}"
    }
  }
}
