package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/netip"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
)

var (
	resourceGroupName = flag.String("g", "", "Name of the created resource group (must be unique).")
	vmName            = flag.String("n", "", "Name of the VM that will run the script provided.")
	nicId             = flag.String("i", "", "azure resource name of the nic (not the full ID).")
	subnetPrefix      = flag.String("s", "0.0.0.0/0", "subnet prefix of the nic to look for")
	//outputFile        *string = flag.String("o", "output-file", "additional output files: tell jogger to write output to a file.")
)

func main() {
	// Replace these with your VM and Resource Group names
	flag.Parse()
	subscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID")
	// Create a new Azure identity
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatalf("failed to obtain an Azure credential: %v", err)
	}
	prefix, err := netip.ParsePrefix(*subnetPrefix)
	if err != nil {
		log.Fatalf("Could not parse provided subnet prefix %s", *subnetPrefix)
	}
	// Create a new network interfaces client
	client, err := armnetwork.NewInterfacesClient(subscriptionID, cred, nil)
	if err != nil {
		log.Fatalf("failed to create network interfaces client: %v", err)
	}

	// List all network interfaces in the resource group
	pager := client.NewListPager(*resourceGroupName, nil)
	for pager.More() {
		page, err := pager.NextPage(context.Background())
		if err != nil {
			log.Fatalf("failed to get next page of network interfaces: %v", err)
		}

		for _, nic := range page.Value {
			// Check if the NIC is attached to the specified VM
			if nic.Properties.VirtualMachine != nil && *nic.Properties.VirtualMachine.ID == fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Compute/virtualMachines/%s", subscriptionID, *resourceGroupName, *vmName) {
				if *nicId == "" || *nic.Name == *nicId {
					for _, ipConfig := range nic.Properties.IPConfigurations {
						address := ipConfig.Properties.PrivateIPAddress
						if address != nil {
							addr, err := netip.ParseAddr(*address)
							if err != nil {
								log.Fatalf("could not parse private ip address: %s", *address)
							}
							if prefix.Contains(addr) {
								fmt.Printf("%s\n", *address)
							}
						}
					}
				}
			}

		}
	}
}
