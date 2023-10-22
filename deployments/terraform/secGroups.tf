resource "openstack_networking_secgroup_v2" "transparencyMonitor" {
  name        = "transparencyMonitor"
  description = "Security group for the transparencyMonitor example instances"
}

resource "openstack_networking_secgroup_rule_v2" "transparencyMonitor" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "icmp"
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = "${openstack_networking_secgroup_v2.transparencyMonitor.id}"
}

resource "openstack_networking_secgroup_rule_v2" "transparencyMonitor_22" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 22
  port_range_max    = 22
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = "${openstack_networking_secgroup_v2.transparencyMonitor.id}"
}

resource "openstack_networking_secgroup_rule_v2" "transparencyMonitor_80" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 80
  port_range_max    = 80
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = "${openstack_networking_secgroup_v2.transparencyMonitor.id}"
}

resource "openstack_networking_secgroup_rule_v2" "transparencyMonitor_7474" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 7474
  port_range_max    = 7474
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = "${openstack_networking_secgroup_v2.transparencyMonitor.id}"
}

resource "openstack_networking_secgroup_rule_v2" "transparencyMonitor_7687" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 7687
  port_range_max    = 7687
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = "${openstack_networking_secgroup_v2.transparencyMonitor.id}"
}

resource "openstack_networking_secgroup_rule_v2" "transparencyMonitor_2376" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 2376
  port_range_max    = 2376
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = "${openstack_networking_secgroup_v2.transparencyMonitor.id}"
}