data "openstack_networking_network_v2" "transparencyMonitor" {
  name = "${var.pool}"
}