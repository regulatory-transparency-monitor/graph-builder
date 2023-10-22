# Monitoring Service
Deploying a TMS on OpenStack VM using Terraform.

## Usage

```sh
# OpenStack access and credentials
$ source ~/openrc.sh 
```

```sh
# Initialize and plan terraform 
$ terraform init
$ terraform plan
```

```sh
# To apply the configurations run 
$ terraform apply
```
```sh
# Destory cluster 
$ terraform destroy
```

## To Access the container UIs:
The FLOATING-IP can be found in terraform/instance_ip.txt or in the output provided be terraform:
```console
openstack_compute_floatingip_associate_v2.transparencyMonitor: Creation complete after 2m46s [id=195.15.194.154/d991abec-8aef-4fab-9698-15126c3a1a40/]
Apply complete! Resources: 17 added, 0 changed, 0 destroyed.

Outputs:

address = "195.15.194.154"
```
 [http://FLOATING-IP:7474](http://FLOATING-IP:7474) access Neo4j database web interface

  [http://FLOATING-IP:5005](http://FLOATING-IP:5005) access the transparency Dashboard

 [http://FLOATING-IP:8080](http://FLOATING-IP:8080) access the GraphQL API


### Conntect to VM

`ssh -i ~/.shh ubuntu@FLOATING-IP`

To stop the containers run:

 `sudo docker compose stop / up`