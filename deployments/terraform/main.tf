resource "openstack_compute_keypair_v2" "transparencyMonitor" {
  name       = "transparencyMonitor"
  public_key = "${file("${var.ssh_key_file}.pub")}"
}

resource "openstack_networking_network_v2" "transparencyMonitor" {
  name           = "transparencyMonitor"
  admin_state_up = "true"
}

resource "openstack_networking_subnet_v2" "transparencyMonitor" {
  name            = "transparencyMonitor"
  network_id      = "${openstack_networking_network_v2.transparencyMonitor.id}"
  cidr            = "10.10.0.0/24"
  ip_version      = 4
  dns_nameservers = ["83.166.143.51", "83.166.143.52"]
}

resource "openstack_networking_router_v2" "transparencyMonitor" {
  name                = "transparencyMonitor"
  admin_state_up      = "true"
  external_network_id = "${data.openstack_networking_network_v2.transparencyMonitor.id}"
}

resource "openstack_networking_router_interface_v2" "transparencyMonitor" {
  router_id = "${openstack_networking_router_v2.transparencyMonitor.id}"
  subnet_id = "${openstack_networking_subnet_v2.transparencyMonitor.id}"
}

resource "openstack_networking_floatingip_v2" "transparencyMonitor" {
  pool = "${var.pool}"
}

resource "openstack_compute_instance_v2" "transparencyMonitor" {
  name            = "transparencyMonitor"
  image_name      = "${var.image}"
  flavor_name     = "${var.flavor}"
  key_pair        = "${openstack_compute_keypair_v2.transparencyMonitor.name}"
  security_groups = ["default", "${openstack_networking_secgroup_v2.transparencyMonitor.name}"]

  network {
    uuid = "${openstack_networking_network_v2.transparencyMonitor.id}"
  }
}

resource "openstack_compute_floatingip_associate_v2" "transparencyMonitor" {
  floating_ip = "${openstack_networking_floatingip_v2.transparencyMonitor.address}"
  instance_id = "${openstack_compute_instance_v2.transparencyMonitor.id}"

  connection {
    host        = "${openstack_networking_floatingip_v2.transparencyMonitor.address}"
    user        = "${var.ssh_user_name}"
    private_key = "${file(var.ssh_key_file)}"
  }

  provisioner "local-exec" {
    command = "echo ${openstack_networking_floatingip_v2.transparencyMonitor.address} > instance_ip.txt"
  }

   provisioner "file" {
    source      = "../../../graph-builder"
    destination = "/home/ubuntu/tms/"
  }

  provisioner "remote-exec" {
    inline = [
      "sudo apt-get -y update",
      "sudo mkfs.ext4 ${openstack_compute_volume_attach_v2.attached.device}",
      "sudo mkdir /mnt/volume",
      "sudo mount ${openstack_compute_volume_attach_v2.attached.device} /mnt/volume",
      "sudo df -h /mnt/volume",
      "sudo mkdir -p /mnt/volume/transparencyMonitor-data",
    ]
  }

  provisioner "remote-exec" {
    inline = [
      "chmod +x /home/ubuntu/tms/deployments/deploy-script.sh",
      "/home/ubuntu/tms/deployments/deploy-script.sh"
    ]
  }

}