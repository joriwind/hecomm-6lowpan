#Hecomm 6lowpan implementation

This implementation cna utilise the hecomm service to connect its nodes to other IoT nodes of IoT networks, not necesairly from the same type.

#Usage
##Define 6LoWPAN interface address
Define the host address which the 6LoWPAN nodes use to communicate with, e.g.:
'''
--address [aaaa::1]:5683
'''

Note: the command "test req" usages the address [::1]:5683 to simulate a node requesting the hecomm network. Use other --address value instead.
