terraform {
  required_providers {
    hidora = {
        source = "HidoraSwiss/hidora"
    }
  }
}

provider "hidora" {
    host = "app.hidora.com"
    access_token = ""
}

data "hidora_create_env" "test-data" {
    id = "env-6378398"
}

resource "hidora_create_env" "test-res" {
    envgroups = "${data.hidora_create_env.test-data.envgroups}" // Can be update
    environment {
        ishaenabled = true
        region = "new" // Can be update
        shortdomain = "env-made-by-terraform" // Force new instance and destroy previous instance !
        sslstate = true
    }
    nodes {
        count = 2
        disklimit = 50
        nodegroup = "cp"
        nodetype = "apache-python"
        env = {
            TEST = "test"
        }
        extip = false
        fixedcloudlets = 4
        flexiblecloudlets = 8
        image = "python"
        mission = "cp"
        tag = "latest"
    }
}