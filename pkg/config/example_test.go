/*
Copyright IBM Corp All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	cb "github.com/hyperledger/fabric-protos-go/common"
	mb "github.com/hyperledger/fabric-protos-go/msp"
	ob "github.com/hyperledger/fabric-protos-go/orderer"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/tools/protolator"
	"github.com/hyperledger/fabric/pkg/config"
	. "github.com/onsi/gomega"
)

const (
	// Arbitrary valid pem encoded x509 certificate from crypto/x509 tests.
	// The contents of the certifcate don't matter, we just need a valid certificate
	// to pass marshaling/unmamarshaling.
	dummyCert = `-----BEGIN CERTIFICATE-----
MIIDATCCAemgAwIBAgIRAKQkkrFx1T/dgB/Go/xBM5swDQYJKoZIhvcNAQELBQAw
EjEQMA4GA1UEChMHQWNtZSBDbzAeFw0xNjA4MTcyMDM2MDdaFw0xNzA4MTcyMDM2
MDdaMBIxEDAOBgNVBAoTB0FjbWUgQ28wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAw
ggEKAoIBAQDAoJtjG7M6InsWwIo+l3qq9u+g2rKFXNu9/mZ24XQ8XhV6PUR+5HQ4
jUFWC58ExYhottqK5zQtKGkw5NuhjowFUgWB/VlNGAUBHtJcWR/062wYrHBYRxJH
qVXOpYKbIWwFKoXu3hcpg/CkdOlDWGKoZKBCwQwUBhWE7MDhpVdQ+ZljUJWL+FlK
yQK5iRsJd5TGJ6VUzLzdT4fmN2DzeK6GLeyMpVpU3sWV90JJbxWQ4YrzkKzYhMmB
EcpXTG2wm+ujiHU/k2p8zlf8Sm7VBM/scmnMFt0ynNXop4FWvJzEm1G0xD2t+e2I
5Utr04dOZPCgkm++QJgYhtZvgW7ZZiGTAgMBAAGjUjBQMA4GA1UdDwEB/wQEAwIF
oDATBgNVHSUEDDAKBggrBgEFBQcDATAMBgNVHRMBAf8EAjAAMBsGA1UdEQQUMBKC
EHRlc3QuZXhhbXBsZS5jb20wDQYJKoZIhvcNAQELBQADggEBADpqKQxrthH5InC7
X96UP0OJCu/lLEMkrjoEWYIQaFl7uLPxKH5AmQPH4lYwF7u7gksR7owVG9QU9fs6
1fK7II9CVgCd/4tZ0zm98FmU4D0lHGtPARrrzoZaqVZcAvRnFTlPX5pFkPhVjjai
/mkxX9LpD8oK1445DFHxK5UjLMmPIIWd8EOi+v5a+hgGwnJpoW7hntSl8kHMtTmy
fnnktsblSUV4lRCit0ymC7Ojhe+gzCCwkgs5kDzVVag+tnl/0e2DloIjASwOhpbH
KVcg7fBd484ht/sS+l0dsB4KDOSpd8JzVDMF8OZqlaydizoJO0yWr9GbCN1+OKq5
EhLrEqU=
-----END CERTIFICATE-----
`

	// Arbitrary valid pem encoded ec private key.
	// The contents of the private key don't matter, we just need a valid
	// EC private key to pass marshaling/unmamarshaling.
	dummyPrivateKey = `-----BEGIN EC PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgDZUgDvKixfLi8cK8
/TFLY97TDmQV3J2ygPpvuI8jSdihRANCAARRN3xgbPIR83dr27UuDaf2OJezpEJx
UC3v06+FD8MUNcRAboqt4akehaNNSh7MMZI+HdnsM4RXN2y8NePUQsPL
-----END EC PRIVATE KEY-----
`

	// Arbitrary valid pem encoded x509 crl.
	// The contents of the CRL don't matter, we just need a valid
	// CRL to pass marshaling/unmamarshaling.
	dummyCRL = `-----BEGIN X509 CRL-----
MIIBYDCBygIBATANBgkqhkiG9w0BAQUFADBDMRMwEQYKCZImiZPyLGQBGRYDY29t
MRcwFQYKCZImiZPyLGQBGRYHZXhhbXBsZTETMBEGA1UEAxMKRXhhbXBsZSBDQRcN
MDUwMjA1MTIwMDAwWhcNMDUwMjA2MTIwMDAwWjAiMCACARIXDTA0MTExOTE1NTcw
M1owDDAKBgNVHRUEAwoBAaAvMC0wHwYDVR0jBBgwFoAUCGivhTPIOUp6+IKTjnBq
SiCELDIwCgYDVR0UBAMCAQwwDQYJKoZIhvcNAQEFBQADgYEAItwYffcIzsx10NBq
m60Q9HYjtIFutW2+DvsVFGzIF20f7pAXom9g5L2qjFXejoRvkvifEBInr0rUL4Xi
NkR9qqNMJTgV/wD9Pn7uPSYS69jnK2LiK8NGgO94gtEVxtCccmrLznrtZ5mLbnCB
fUNCdMGmr8FVF6IzTNYGmCuk/C4=
-----END X509 CRL-----
`
)

func Example_systemChannel() {
	baseConfig := fetchSystemChannelConfig()
	c := config.New(baseConfig)

	err := c.UpdateConsortiumChannelCreationPolicy("SampleConsortium",
		config.Policy{Type: config.ImplicitMetaPolicyType, Rule: "MAJORITY Admins"})
	if err != nil {
		panic(err)
	}

	err = c.RemoveConsortium("SampleConsortium2")
	if err != nil {
		panic(err)
	}

	orgToAdd := config.Organization{
		Name: "Org3",
		Policies: map[string]config.Policy{
			config.AdminsPolicyKey: {
				Type: config.ImplicitMetaPolicyType,
				Rule: "MAJORITY Admins",
			},
			config.EndorsementPolicyKey: {
				Type: config.ImplicitMetaPolicyType,
				Rule: "MAJORITY Endorsement",
			},
			config.ReadersPolicyKey: {
				Type: config.ImplicitMetaPolicyType,
				Rule: "ANY Readers",
			},
			config.WritersPolicyKey: {
				Type: config.ImplicitMetaPolicyType,
				Rule: "ANY Writers",
			},
		},
		MSP: baseMSP(&testing.T{}),
	}

	err = c.AddOrgToConsortium(orgToAdd, "SampleConsortium")
	if err != nil {
		panic(err)
	}

	err = c.RemoveConsortiumOrg("SampleConsortium", "Org3")
	if err != nil {
		panic(err)
	}

	// Compute the delta
	configUpdate, err := c.ComputeUpdate("testsyschannel")
	if err != nil {
		panic(err)
	}

	// Collect the necessary signatures
	// The example respresents a 2 peer 1 org channel, to meet the policies defined
	// the transaction will be signed by both peers
	configSignatures := []*cb.ConfigSignature{}

	peer1SigningIdentity := createSigningIdentity()
	peer2SigningIdentity := createSigningIdentity()

	signingIdentities := []config.SigningIdentity{
		peer1SigningIdentity,
		peer2SigningIdentity,
	}

	for _, si := range signingIdentities {
		// Sign the config update with the specified signer identity
		configSignature, err := si.SignConfigUpdate(configUpdate)
		if err != nil {
			panic(err)
		}

		configSignatures = append(configSignatures, configSignature)
	}

	// Sign the envelope with the list of signatures
	envelope, err := peer1SigningIdentity.SignConfigUpdateEnvelope(configUpdate, configSignatures...)
	if err != nil {
		panic(err)
	}

	// The below logic outputs the signed envelope in JSON format

	// The timestamps of the ChannelHeader varies so this comparison only considers the ConfigUpdateEnvelope JSON.
	payload := &cb.Payload{}

	err = proto.Unmarshal(envelope.Payload, payload)
	if err != nil {
		panic(err)
	}

	data := &cb.ConfigUpdateEnvelope{}

	err = proto.Unmarshal(payload.Data, data)
	if err != nil {
		panic(err)
	}

	// Signature and nonce is different on every example run
	data.Signatures = nil

	err = protolator.DeepMarshalJSON(os.Stdout, data)
	if err != nil {
		panic(err)
	}

	// Output:
	// {
	//	"config_update": {
	//		"channel_id": "testsyschannel",
	//		"isolated_data": {},
	//		"read_set": {
	//			"groups": {
	//				"Consortiums": {
	//					"groups": {
	//						"SampleConsortium": {
	//							"groups": {},
	//							"mod_policy": "",
	//							"policies": {},
	//							"values": {},
	//							"version": "0"
	//						}
	//					},
	//					"mod_policy": "",
	//					"policies": {},
	//					"values": {},
	//					"version": "0"
	//				}
	//			},
	//			"mod_policy": "",
	//			"policies": {},
	//			"values": {},
	//			"version": "0"
	//		},
	//		"write_set": {
	//			"groups": {
	//				"Consortiums": {
	//					"groups": {
	//						"SampleConsortium": {
	//							"groups": {},
	//							"mod_policy": "",
	//							"policies": {},
	//							"values": {
	//								"ChannelCreationPolicy": {
	//									"mod_policy": "/Channel/Orderer/Admins",
	//									"value": {
	//										"type": 3,
	//										"value": {
	//											"rule": "MAJORITY",
	//											"sub_policy": "Admins"
	//										}
	//									},
	//									"version": "1"
	//								}
	//							},
	//							"version": "0"
	//						}
	//					},
	//					"mod_policy": "",
	//					"policies": {},
	//					"values": {},
	//					"version": "1"
	//				}
	//			},
	//			"mod_policy": "",
	//			"policies": {},
	//			"values": {},
	//			"version": "0"
	//		}
	//	},
	//	"signatures": []
	// }
}

func Example_orderer() {
	baseConfig := fetchChannelConfig()
	c := config.New(baseConfig)

	// Must retrieve the current orderer configuration from block and modify
	// the desired values
	orderer, err := c.OrdererConfiguration()
	if err != nil {
		panic(err)
	}

	orderer.Kafka.Brokers = []string{"kafka0:9092", "kafka1:9092", "kafka2:9092"}
	orderer.BatchSize.MaxMessageCount = 500

	err = c.UpdateOrdererConfiguration(orderer)
	if err != nil {
		panic(nil)
	}

	err = c.RemoveOrdererPolicy(config.WritersPolicyKey)
	if err != nil {
		panic(err)
	}

	err = c.AddOrdererPolicy(config.AdminsPolicyKey, "TestPolicy", config.Policy{
		Type: config.ImplicitMetaPolicyType,
		Rule: "MAJORITY Endorsement",
	})
	if err != nil {
		panic(err)
	}

}

func Example_application() {
	baseConfig := fetchChannelConfig()
	c := config.New(baseConfig)

	acls := map[string]string{
		"peer/Propose": "/Channel/Application/Writers",
	}

	err := c.AddACLs(acls)
	if err != nil {
		panic(err)
	}

	aclsToDelete := []string{"event/Block"}

	err = c.RemoveACLs(aclsToDelete)
	if err != nil {
		panic(err)
	}

	err = c.AddApplicationPolicy(config.AdminsPolicyKey, "TestPolicy", config.Policy{
		Type: config.ImplicitMetaPolicyType,
		Rule: "MAJORITY Endorsement",
	})
	if err != nil {
		panic(err)
	}
}

func Example_organization() {
	baseConfig := fetchChannelConfig()
	c := config.New(baseConfig)

	// Application Organization
	newAnchorPeer := config.Address{
		Host: "127.0.0.2",
		Port: 7051,
	}

	// Add a new anchor peer
	err := c.AddAnchorPeer("Org1", newAnchorPeer)
	if err != nil {
		panic(err)
	}

	oldAnchorPeer := config.Address{
		Host: "127.0.0.1",
		Port: 7051,
	}

	// Remove an anchor peer from Org1
	err = c.RemoveAnchorPeer("Org1", oldAnchorPeer)
	if err != nil {
		panic(err)
	}

	appOrg := config.Organization{
		Name: "Org2",
		MSP:  baseMSP(&testing.T{}),
		Policies: map[string]config.Policy{
			config.AdminsPolicyKey: {
				Type: config.ImplicitMetaPolicyType,
				Rule: "MAJORITY Admins",
			},
			config.EndorsementPolicyKey: {
				Type: config.ImplicitMetaPolicyType,
				Rule: "MAJORITY Endorsement",
			},
			config.LifecycleEndorsementPolicyKey: {
				Type: config.ImplicitMetaPolicyType,
				Rule: "MAJORITY Endorsement",
			},
			config.ReadersPolicyKey: {
				Type: config.ImplicitMetaPolicyType,
				Rule: "ANY Readers",
			},
			config.WritersPolicyKey: {
				Type: config.ImplicitMetaPolicyType,
				Rule: "ANY Writers",
			},
		},
		AnchorPeers: []config.Address{
			{
				Host: "127.0.0.1",
				Port: 7051,
			},
		},
	}

	err = c.AddApplicationOrg(appOrg)
	if err != nil {
		panic(err)
	}

	err = c.RemoveApplicationOrg("Org2")
	if err != nil {
		panic(err)
	}

	err = c.AddApplicationOrgPolicy("Org1", config.AdminsPolicyKey, "TestPolicy", config.Policy{
		Type: config.ImplicitMetaPolicyType,
		Rule: "MAJORITY Endorsement",
	})
	if err != nil {
		panic(err)
	}

	err = c.RemoveApplicationOrgPolicy("Org1", config.WritersPolicyKey)
	if err != nil {
		panic(err)
	}

	// Orderer Organization
	ordererOrg := appOrg
	ordererOrg.Name = "OrdererOrg2"
	ordererOrg.AnchorPeers = nil

	err = c.AddOrdererOrg(ordererOrg)
	if err != nil {
		panic(err)
	}

	err = c.RemoveOrdererOrg("OrdererOrg2")
	if err != nil {
		panic(err)
	}

	err = c.RemoveOrdererOrgPolicy("OrdererOrg", config.WritersPolicyKey)
	if err != nil {
		panic(err)
	}

	err = c.AddOrdererOrgPolicy("OrdererOrg", config.AdminsPolicyKey, "TestPolicy", config.Policy{
		Type: config.ImplicitMetaPolicyType,
		Rule: "MAJORITY Endorsement",
	})
	if err != nil {
		panic(err)
	}

	err = c.AddOrdererEndpoint("OrdererOrg", config.Address{Host: "127.0.0.3", Port: 8050})
	if err != nil {
		panic(err)
	}

	err = c.RemoveOrdererEndpoint("OrdererOrg", config.Address{Host: "127.0.0.1", Port: 9050})
	if err != nil {
		panic(err)
	}
}

func ExampleNewCreateChannelTx() {
	channel := config.Channel{
		Consortium: "SampleConsortium",
		Application: config.Application{
			Organizations: []config.Organization{
				{
					Name: "Org1",
				},
				{
					Name: "Org2",
				},
			},
			Capabilities: []string{"V1_3"},
			ACLs:         map[string]string{"event/Block": "/Channel/Application/Readers"},
			Policies: map[string]config.Policy{
				config.ReadersPolicyKey: {
					Type: config.ImplicitMetaPolicyType,
					Rule: "ANY Readers",
				},
				config.WritersPolicyKey: {
					Type: config.ImplicitMetaPolicyType,
					Rule: "ANY Writers",
				},
				config.AdminsPolicyKey: {
					Type: config.ImplicitMetaPolicyType,
					Rule: "MAJORITY Admins",
				},
				config.EndorsementPolicyKey: {
					Type: config.ImplicitMetaPolicyType,
					Rule: "MAJORITY Endorsement",
				},
				config.LifecycleEndorsementPolicyKey: {
					Type: config.ImplicitMetaPolicyType,
					Rule: "MAJORITY Endorsement",
				},
			},
		},
	}
	channelID := "testchannel"
	envelope, err := config.NewCreateChannelTx(channel, channelID)
	if err != nil {
		panic(err)
	}

	// The timestamps of the ChannelHeader varies so this comparison only considers the ConfigUpdateEnvelope JSON.
	payload := &cb.Payload{}

	err = proto.Unmarshal(envelope.Payload, payload)
	if err != nil {
		panic(err)
	}

	data := &cb.ConfigUpdateEnvelope{}

	err = proto.Unmarshal(payload.Data, data)
	if err != nil {
		panic(err)
	}

	err = protolator.DeepMarshalJSON(os.Stdout, data)
	if err != nil {
		panic(err)
	}

	// Output:
	// {
	// 	"config_update": {
	// 		"channel_id": "testchannel",
	// 		"isolated_data": {},
	// 		"read_set": {
	// 			"groups": {
	// 				"Application": {
	// 					"groups": {
	// 						"Org1": {
	// 							"groups": {},
	// 							"mod_policy": "",
	// 							"policies": {},
	// 							"values": {},
	// 							"version": "0"
	// 						},
	// 						"Org2": {
	// 							"groups": {},
	// 							"mod_policy": "",
	// 							"policies": {},
	// 							"values": {},
	// 							"version": "0"
	// 						}
	// 					},
	// 					"mod_policy": "",
	// 					"policies": {},
	// 					"values": {},
	// 					"version": "0"
	// 				}
	// 			},
	// 			"mod_policy": "",
	// 			"policies": {},
	// 			"values": {
	// 				"Consortium": {
	// 					"mod_policy": "",
	// 					"value": null,
	// 					"version": "0"
	// 				}
	// 			},
	// 			"version": "0"
	// 		},
	// 		"write_set": {
	// 			"groups": {
	// 				"Application": {
	// 					"groups": {
	// 						"Org1": {
	// 							"groups": {},
	// 							"mod_policy": "",
	// 							"policies": {},
	// 							"values": {},
	// 							"version": "0"
	// 						},
	// 						"Org2": {
	// 							"groups": {},
	// 							"mod_policy": "",
	// 							"policies": {},
	// 							"values": {},
	// 							"version": "0"
	// 						}
	// 					},
	// 					"mod_policy": "Admins",
	// 					"policies": {
	// 						"Admins": {
	// 							"mod_policy": "Admins",
	// 							"policy": {
	// 								"type": 3,
	// 								"value": {
	// 									"rule": "MAJORITY",
	// 									"sub_policy": "Admins"
	// 								}
	// 							},
	// 							"version": "0"
	// 						},
	// 						"Endorsement": {
	// 							"mod_policy": "Admins",
	// 							"policy": {
	// 								"type": 3,
	// 								"value": {
	// 									"rule": "MAJORITY",
	// 									"sub_policy": "Endorsement"
	// 								}
	// 							},
	// 							"version": "0"
	// 						},
	// 						"LifecycleEndorsement": {
	// 							"mod_policy": "Admins",
	// 							"policy": {
	// 								"type": 3,
	// 								"value": {
	// 									"rule": "MAJORITY",
	// 									"sub_policy": "Endorsement"
	// 								}
	// 							},
	// 							"version": "0"
	// 						},
	// 						"Readers": {
	// 							"mod_policy": "Admins",
	// 							"policy": {
	// 								"type": 3,
	// 								"value": {
	// 									"rule": "ANY",
	// 									"sub_policy": "Readers"
	// 								}
	// 							},
	// 							"version": "0"
	// 						},
	// 						"Writers": {
	// 							"mod_policy": "Admins",
	// 							"policy": {
	// 								"type": 3,
	// 								"value": {
	// 									"rule": "ANY",
	// 									"sub_policy": "Writers"
	// 								}
	// 							},
	// 							"version": "0"
	// 						}
	// 					},
	// 					"values": {
	// 						"ACLs": {
	// 							"mod_policy": "Admins",
	// 							"value": {
	// 								"acls": {
	// 									"event/Block": {
	// 										"policy_ref": "/Channel/Application/Readers"
	// 									}
	// 								}
	// 							},
	// 							"version": "0"
	// 						},
	// 						"Capabilities": {
	// 							"mod_policy": "Admins",
	// 							"value": {
	// 								"capabilities": {
	// 									"V1_3": {}
	// 								}
	// 							},
	// 							"version": "0"
	// 						}
	// 					},
	// 					"version": "1"
	// 				}
	// 			},
	// 			"mod_policy": "",
	// 			"policies": {},
	// 			"values": {
	// 				"Consortium": {
	// 					"mod_policy": "",
	// 					"value": {
	// 						"name": "SampleConsortium"
	// 					},
	// 					"version": "0"
	// 				}
	// 			},
	// 			"version": "0"
	// 		}
	// 	},
	// 	"signatures": []
	// }
}

func ExampleNew() {
	baseConfig := fetchChannelConfig()
	_ = config.New(baseConfig)
}

func ExampleConfigTx_AddChannelCapability() {
	baseConfig := fetchChannelConfig()
	c := config.New(baseConfig)

	err := c.AddChannelCapability("V1_3")
	if err != nil {
		panic(err)
	}

	err = protolator.DeepMarshalJSON(os.Stdout, c.Updated())
	if err != nil {
		panic(err)
	}

	// Output:
	// {
	// 	"channel_group": {
	// 		"groups": {
	// 			"Application": {
	// 				"groups": {
	// 					"Org1": {
	// 						"groups": {},
	// 						"mod_policy": "",
	// 						"policies": {},
	// 						"values": {
	// 							"AnchorPeers": {
	// 								"mod_policy": "Admins",
	// 								"value": {
	// 									"anchor_peers": [
	// 										{
	// 											"host": "127.0.0.1",
	// 											"port": 7050
	// 										}
	// 									]
	// 								},
	// 								"version": "0"
	// 							},
	// 							"MSP": {
	// 								"mod_policy": "Admins",
	// 								"value": null,
	// 								"version": "0"
	// 							}
	// 						},
	// 						"version": "0"
	// 					}
	// 				},
	// 				"mod_policy": "",
	// 				"policies": {
	// 					"Admins": {
	// 						"mod_policy": "Admins",
	// 						"policy": {
	// 							"type": 3,
	// 							"value": {
	// 								"rule": "MAJORITY",
	// 								"sub_policy": "Admins"
	// 							}
	// 						},
	// 						"version": "0"
	// 					},
	// 					"LifecycleEndorsement": {
	// 						"mod_policy": "Admins",
	// 						"policy": {
	// 							"type": 3,
	// 							"value": {
	// 								"rule": "MAJORITY",
	// 								"sub_policy": "Admins"
	// 							}
	// 						},
	// 						"version": "0"
	// 					},
	// 					"Readers": {
	// 						"mod_policy": "Admins",
	// 						"policy": {
	// 							"type": 3,
	// 							"value": {
	// 								"rule": "ANY",
	// 								"sub_policy": "Readers"
	// 							}
	// 						},
	// 						"version": "0"
	// 					},
	// 					"Writers": {
	// 						"mod_policy": "Admins",
	// 						"policy": {
	// 							"type": 3,
	// 							"value": {
	// 								"rule": "ANY",
	// 								"sub_policy": "Writers"
	// 							}
	// 						},
	// 						"version": "0"
	// 					}
	// 				},
	// 				"values": {
	// 					"ACLs": {
	// 						"mod_policy": "Admins",
	// 						"value": {
	// 							"acls": {
	// 								"event/block": {
	// 									"policy_ref": "/Channel/Application/Readers"
	// 								}
	// 							}
	// 						},
	// 						"version": "0"
	// 					},
	// 					"Capabilities": {
	// 						"mod_policy": "Admins",
	// 						"value": {
	// 							"capabilities": {
	// 								"V1_3": {}
	// 							}
	// 						},
	// 						"version": "0"
	// 					}
	// 				},
	// 				"version": "0"
	// 			},
	// 			"Orderer": {
	// 				"groups": {
	// 					"OrdererOrg": {
	// 						"groups": {},
	// 						"mod_policy": "Admins",
	// 						"policies": {
	// 							"Admins": {
	// 								"mod_policy": "Admins",
	// 								"policy": {
	// 									"type": 3,
	// 									"value": {
	// 										"rule": "MAJORITY",
	// 										"sub_policy": "Admins"
	// 									}
	// 								},
	// 								"version": "0"
	// 							},
	// 							"Readers": {
	// 								"mod_policy": "Admins",
	// 								"policy": {
	// 									"type": 3,
	// 									"value": {
	// 										"rule": "ANY",
	// 										"sub_policy": "Readers"
	// 									}
	// 								},
	// 								"version": "0"
	// 							},
	// 							"Writers": {
	// 								"mod_policy": "Admins",
	// 								"policy": {
	// 									"type": 3,
	// 									"value": {
	// 										"rule": "ANY",
	// 										"sub_policy": "Writers"
	// 									}
	// 								},
	// 								"version": "0"
	// 							}
	// 						},
	// 						"values": {
	// 							"Endpoints": {
	// 								"mod_policy": "Admins",
	// 								"value": {
	// 									"addresses": [
	// 										"127.0.0.1:7050"
	// 									]
	// 								},
	// 								"version": "0"
	// 							},
	// 							"MSP": {
	// 								"mod_policy": "Admins",
	// 								"value": null,
	// 								"version": "0"
	// 							}
	// 						},
	// 						"version": "0"
	// 					}
	// 				},
	// 				"mod_policy": "",
	// 				"policies": {
	// 					"Admins": {
	// 						"mod_policy": "Admins",
	// 						"policy": {
	// 							"type": 3,
	// 							"value": {
	// 								"rule": "MAJORITY",
	// 								"sub_policy": "Admins"
	// 							}
	// 						},
	// 						"version": "0"
	// 					},
	// 					"BlockValidation": {
	// 						"mod_policy": "Admins",
	// 						"policy": {
	// 							"type": 3,
	// 							"value": {
	// 								"rule": "ANY",
	// 								"sub_policy": "Writers"
	// 							}
	// 						},
	// 						"version": "0"
	// 					},
	// 					"Readers": {
	// 						"mod_policy": "Admins",
	// 						"policy": {
	// 							"type": 3,
	// 							"value": {
	// 								"rule": "ANY",
	// 								"sub_policy": "Readers"
	// 							}
	// 						},
	// 						"version": "0"
	// 					},
	// 					"Writers": {
	// 						"mod_policy": "Admins",
	// 						"policy": {
	// 							"type": 3,
	// 							"value": {
	// 								"rule": "ANY",
	// 								"sub_policy": "Writers"
	// 							}
	// 						},
	// 						"version": "0"
	// 					}
	// 				},
	// 				"values": {
	// 					"BatchSize": {
	// 						"mod_policy": "",
	// 						"value": {
	// 							"absolute_max_bytes": 100,
	// 							"max_message_count": 100,
	// 							"preferred_max_bytes": 100
	// 						},
	// 						"version": "0"
	// 					},
	// 					"BatchTimeout": {
	// 						"mod_policy": "",
	// 						"value": {
	// 							"timeout": "15s"
	// 						},
	// 						"version": "0"
	// 					},
	// 					"Capabilities": {
	// 						"mod_policy": "Admins",
	// 						"value": {
	// 							"capabilities": {
	// 								"V1_3": {}
	// 							}
	// 						},
	// 						"version": "0"
	// 					},
	// 					"ChannelRestrictions": {
	// 						"mod_policy": "Admins",
	// 						"value": {
	// 							"max_count": "1"
	// 						},
	// 						"version": "0"
	// 					},
	// 					"ConsensusType": {
	// 						"mod_policy": "Admins",
	// 						"value": {
	// 							"metadata": null,
	// 							"state": "STATE_NORMAL",
	// 							"type": "kafka"
	// 						},
	// 						"version": "0"
	// 					},
	// 					"KafkaBrokers": {
	// 						"mod_policy": "Admins",
	// 						"value": {
	// 							"brokers": [
	// 								"kafka0:9092",
	// 								"kafka1:9092"
	// 							]
	// 						},
	// 						"version": "0"
	// 					}
	// 				},
	// 				"version": "1"
	// 			}
	// 		},
	// 		"mod_policy": "",
	// 		"policies": {
	// 			"Admins": {
	// 				"mod_policy": "Admins",
	// 				"policy": {
	// 					"type": 3,
	// 					"value": {
	// 						"rule": "MAJORITY",
	// 						"sub_policy": "Admins"
	// 					}
	// 				},
	// 				"version": "0"
	// 			},
	// 			"Readers": {
	// 				"mod_policy": "Admins",
	// 				"policy": {
	// 					"type": 3,
	// 					"value": {
	// 						"rule": "ANY",
	// 						"sub_policy": "Readers"
	// 					}
	// 				},
	// 				"version": "0"
	// 			},
	// 			"Writers": {
	// 				"mod_policy": "Admins",
	// 				"policy": {
	// 					"type": 3,
	// 					"value": {
	// 						"rule": "ANY",
	// 						"sub_policy": "Writers"
	// 					}
	// 				},
	// 				"version": "0"
	// 			}
	// 		},
	// 		"values": {
	// 			"Capabilities": {
	// 				"mod_policy": "Admins",
	// 				"value": {
	// 					"capabilities": {
	// 						"V1_3": {}
	// 					}
	// 				},
	// 				"version": "0"
	// 			},
	// 			"OrdererAddresses": {
	// 				"mod_policy": "Admins",
	// 				"value": {
	// 					"addresses": [
	// 						"127.0.0.1:7050"
	// 					]
	// 				},
	// 				"version": "0"
	// 			}
	// 		},
	// 		"version": "0"
	// 	},
	// 	"sequence": "0"
	// }
}

func ExampleConfigTx_AddOrdererCapability() {
	baseConfig := fetchChannelConfig()
	c := config.New(baseConfig)

	err := c.AddOrdererCapability("V1_4")
	if err != nil {
		panic(err)
	}
}

func ExampleConfigTx_AddApplicationCapability() {
	baseConfig := fetchChannelConfig()
	c := config.New(baseConfig)

	err := c.AddChannelCapability("V1_3")
	if err != nil {
		panic(err)
	}
}

func ExampleConfigTx_RemoveChannelCapability() {
	baseConfig := fetchChannelConfig()
	c := config.New(baseConfig)

	err := c.RemoveChannelCapability("V1_3")
	if err != nil {
		panic(err)
	}
}

func ExampleConfigTx_RemoveOrdererCapability() {
	baseConfig := fetchChannelConfig()
	c := config.New(baseConfig)

	err := c.RemoveOrdererCapability("V1_4")
	if err != nil {
		panic(err)
	}
}

func ExampleConfigTx_UpdateApplicationMSP() {
	baseConfig := fetchChannelConfig()
	c := config.New(baseConfig)

	msp, err := c.ApplicationMSP("Org1")
	if err != nil {
		panic(err)
	}

	newIntermediateCert := &x509.Certificate{
		KeyUsage: x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		IsCA:     true,
	}

	msp.IntermediateCerts = append(msp.IntermediateCerts, newIntermediateCert)

	err = c.UpdateApplicationMSP(msp, "Org1")
	if err != nil {
		panic(err)
	}
}

func ExampleConfigTx_RemoveApplicationCapability() {
	baseConfig := fetchChannelConfig()
	c := config.New(baseConfig)

	err := c.RemoveChannelCapability("V1_3")
	if err != nil {
		panic(err)
	}
}

// fetchChannelConfig mocks retrieving the config transaction from the most recent configuration block.
func fetchChannelConfig() *cb.Config {
	return &cb.Config{
		ChannelGroup: &cb.ConfigGroup{
			Groups: map[string]*cb.ConfigGroup{
				config.OrdererGroupKey: {
					Version: 1,
					Groups: map[string]*cb.ConfigGroup{
						"OrdererOrg": {
							Groups: map[string]*cb.ConfigGroup{},
							Values: map[string]*cb.ConfigValue{
								config.EndpointsKey: {
									ModPolicy: config.AdminsPolicyKey,
									Value: marshalOrPanic(&cb.OrdererAddresses{
										Addresses: []string{"127.0.0.1:7050"},
									}),
								},
								config.MSPKey: {
									ModPolicy: config.AdminsPolicyKey,
									Value: marshalOrPanic(&mb.MSPConfig{
										Config: []byte{},
									}),
								},
							},
							Policies: map[string]*cb.ConfigPolicy{
								config.AdminsPolicyKey: {
									ModPolicy: config.AdminsPolicyKey,
									Policy: &cb.Policy{
										Type: 3,
										Value: marshalOrPanic(&cb.ImplicitMetaPolicy{
											Rule:      cb.ImplicitMetaPolicy_MAJORITY,
											SubPolicy: config.AdminsPolicyKey,
										}),
									},
								},
								config.ReadersPolicyKey: {
									ModPolicy: config.AdminsPolicyKey,
									Policy: &cb.Policy{
										Type: 3,
										Value: marshalOrPanic(&cb.ImplicitMetaPolicy{
											Rule:      cb.ImplicitMetaPolicy_ANY,
											SubPolicy: config.ReadersPolicyKey,
										}),
									},
								},
								config.WritersPolicyKey: {
									ModPolicy: config.AdminsPolicyKey,
									Policy: &cb.Policy{
										Type: 3,
										Value: marshalOrPanic(&cb.ImplicitMetaPolicy{
											Rule:      cb.ImplicitMetaPolicy_ANY,
											SubPolicy: config.WritersPolicyKey,
										}),
									},
								},
							},
							ModPolicy: config.AdminsPolicyKey,
						},
					},
					Values: map[string]*cb.ConfigValue{
						config.ConsensusTypeKey: {
							ModPolicy: config.AdminsPolicyKey,
							Value: marshalOrPanic(&ob.ConsensusType{
								Type: config.ConsensusTypeKafka,
							}),
						},
						config.ChannelRestrictionsKey: {
							ModPolicy: config.AdminsPolicyKey,
							Value: marshalOrPanic(&ob.ChannelRestrictions{
								MaxCount: 1,
							}),
						},
						config.CapabilitiesKey: {
							ModPolicy: config.AdminsPolicyKey,
							Value: marshalOrPanic(&cb.Capabilities{
								Capabilities: map[string]*cb.Capability{
									"V1_3": {},
								},
							}),
						},
						config.KafkaBrokersKey: {
							ModPolicy: config.AdminsPolicyKey,
							Value: marshalOrPanic(&ob.KafkaBrokers{
								Brokers: []string{"kafka0:9092", "kafka1:9092"},
							}),
						},
						config.BatchTimeoutKey: {
							Value: marshalOrPanic(&ob.BatchTimeout{
								Timeout: "15s",
							}),
						},
						config.BatchSizeKey: {
							Value: marshalOrPanic(&ob.BatchSize{
								MaxMessageCount:   100,
								AbsoluteMaxBytes:  100,
								PreferredMaxBytes: 100,
							}),
						},
					},
					Policies: map[string]*cb.ConfigPolicy{
						config.AdminsPolicyKey: {
							ModPolicy: config.AdminsPolicyKey,
							Policy: &cb.Policy{
								Type: 3,
								Value: marshalOrPanic(&cb.ImplicitMetaPolicy{
									Rule:      cb.ImplicitMetaPolicy_MAJORITY,
									SubPolicy: config.AdminsPolicyKey,
								}),
							},
						},
						config.ReadersPolicyKey: {
							ModPolicy: config.AdminsPolicyKey,
							Policy: &cb.Policy{
								Type: 3,
								Value: marshalOrPanic(&cb.ImplicitMetaPolicy{
									Rule:      cb.ImplicitMetaPolicy_ANY,
									SubPolicy: config.ReadersPolicyKey,
								}),
							},
						},
						config.WritersPolicyKey: {
							ModPolicy: config.AdminsPolicyKey,
							Policy: &cb.Policy{
								Type: 3,
								Value: marshalOrPanic(&cb.ImplicitMetaPolicy{
									Rule:      cb.ImplicitMetaPolicy_ANY,
									SubPolicy: config.WritersPolicyKey,
								}),
							},
						},
						config.BlockValidationPolicyKey: {
							ModPolicy: config.AdminsPolicyKey,
							Policy: &cb.Policy{
								Type: 3,
								Value: marshalOrPanic(&cb.ImplicitMetaPolicy{
									Rule:      cb.ImplicitMetaPolicy_ANY,
									SubPolicy: config.WritersPolicyKey,
								}),
							},
						},
					},
				},
				config.ApplicationGroupKey: {
					Groups: map[string]*cb.ConfigGroup{
						"Org1": {
							Groups: map[string]*cb.ConfigGroup{},
							Values: map[string]*cb.ConfigValue{
								config.AnchorPeersKey: {
									ModPolicy: config.AdminsPolicyKey,
									Value: marshalOrPanic(&pb.AnchorPeers{
										AnchorPeers: []*pb.AnchorPeer{
											{Host: "127.0.0.1", Port: 7050},
										},
									}),
								},
								config.MSPKey: {
									ModPolicy: config.AdminsPolicyKey,
									Value: marshalOrPanic(&mb.MSPConfig{
										Config: []byte{},
									}),
								},
							},
						},
					},
					Values: map[string]*cb.ConfigValue{
						config.ACLsKey: {
							ModPolicy: config.AdminsPolicyKey,
							Value: marshalOrPanic(&pb.ACLs{
								Acls: map[string]*pb.APIResource{
									"event/block": {PolicyRef: "/Channel/Application/Readers"},
								},
							}),
						},
						config.CapabilitiesKey: {
							ModPolicy: config.AdminsPolicyKey,
							Value: marshalOrPanic(&cb.Capabilities{
								Capabilities: map[string]*cb.Capability{
									"V1_3": {},
								},
							}),
						},
					},
					Policies: map[string]*cb.ConfigPolicy{
						config.LifecycleEndorsementPolicyKey: {
							ModPolicy: config.AdminsPolicyKey,
							Policy: &cb.Policy{
								Type: 3,
								Value: marshalOrPanic(&cb.ImplicitMetaPolicy{
									Rule:      cb.ImplicitMetaPolicy_MAJORITY,
									SubPolicy: config.AdminsPolicyKey,
								}),
							},
						},
						config.AdminsPolicyKey: {
							ModPolicy: config.AdminsPolicyKey,
							Policy: &cb.Policy{
								Type: 3,
								Value: marshalOrPanic(&cb.ImplicitMetaPolicy{
									Rule:      cb.ImplicitMetaPolicy_MAJORITY,
									SubPolicy: config.AdminsPolicyKey,
								}),
							},
						},
						config.ReadersPolicyKey: {
							ModPolicy: config.AdminsPolicyKey,
							Policy: &cb.Policy{
								Type: 3,
								Value: marshalOrPanic(&cb.ImplicitMetaPolicy{
									Rule:      cb.ImplicitMetaPolicy_ANY,
									SubPolicy: config.ReadersPolicyKey,
								}),
							},
						},
						config.WritersPolicyKey: {
							ModPolicy: config.AdminsPolicyKey,
							Policy: &cb.Policy{
								Type: 3,
								Value: marshalOrPanic(&cb.ImplicitMetaPolicy{
									Rule:      cb.ImplicitMetaPolicy_ANY,
									SubPolicy: config.WritersPolicyKey,
								}),
							},
						},
					},
				},
			},
			Values: map[string]*cb.ConfigValue{
				config.OrdererAddressesKey: {
					Value: marshalOrPanic(&cb.OrdererAddresses{
						Addresses: []string{"127.0.0.1:7050"},
					}),
					ModPolicy: config.AdminsPolicyKey,
				},
			},
			Policies: map[string]*cb.ConfigPolicy{
				config.AdminsPolicyKey: {
					ModPolicy: config.AdminsPolicyKey,
					Policy: &cb.Policy{
						Type: 3,
						Value: marshalOrPanic(&cb.ImplicitMetaPolicy{
							Rule:      cb.ImplicitMetaPolicy_MAJORITY,
							SubPolicy: config.AdminsPolicyKey,
						}),
					},
				},
				config.ReadersPolicyKey: {
					ModPolicy: config.AdminsPolicyKey,
					Policy: &cb.Policy{
						Type: 3,
						Value: marshalOrPanic(&cb.ImplicitMetaPolicy{
							Rule:      cb.ImplicitMetaPolicy_ANY,
							SubPolicy: config.ReadersPolicyKey,
						}),
					},
				},
				config.WritersPolicyKey: {
					ModPolicy: config.AdminsPolicyKey,
					Policy: &cb.Policy{
						Type: 3,
						Value: marshalOrPanic(&cb.ImplicitMetaPolicy{
							Rule:      cb.ImplicitMetaPolicy_ANY,
							SubPolicy: config.WritersPolicyKey,
						}),
					},
				},
			},
		},
	}
}

// marshalOrPanic is a helper for proto marshal.
func marshalOrPanic(pb proto.Message) []byte {
	data, err := proto.Marshal(pb)
	if err != nil {
		panic(err)
	}

	return data
}

// createSigningIdentity returns a identity that can be used for signing transactions.
// Signing identity can be retrieved from MSP configuration for each peer.
func createSigningIdentity() config.SigningIdentity {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(fmt.Sprintf("Failed to generate private key: %v", err))
	}

	return config.SigningIdentity{
		Certificate: generateCert(),
		PrivateKey:  privKey,
		MSPID:       "Org1MSP",
	}
}

// generateCert creates a certificate for the SigningIdentity.
func generateCert() *x509.Certificate {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)

	if err != nil {
		log.Fatalf("Failed to generate serial number: %s", err)
	}

	return &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   "Wile E. Coyote",
			Organization: []string{"Acme Co"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
}

func fetchSystemChannelConfig() *cb.Config {
	return &cb.Config{
		ChannelGroup: &cb.ConfigGroup{
			Groups: map[string]*cb.ConfigGroup{
				config.ConsortiumsGroupKey: {
					Groups: map[string]*cb.ConfigGroup{
						"SampleConsortium": {
							Groups: map[string]*cb.ConfigGroup{
								"Org1": {
									Groups:    map[string]*cb.ConfigGroup{},
									Policies:  map[string]*cb.ConfigPolicy{},
									Values:    map[string]*cb.ConfigValue{},
									ModPolicy: "Admins",
									Version:   0,
								},
								"Org2": {
									Groups:    map[string]*cb.ConfigGroup{},
									Policies:  map[string]*cb.ConfigPolicy{},
									Values:    map[string]*cb.ConfigValue{},
									ModPolicy: "Admins",
									Version:   0,
								},
							},
							Values: map[string]*cb.ConfigValue{
								config.ChannelCreationPolicyKey: {
									ModPolicy: "/Channel/Orderer/Admins",
									Value: marshalOrPanic(&cb.Policy{
										Type: 3,
										Value: marshalOrPanic(&cb.ImplicitMetaPolicy{
											Rule:      cb.ImplicitMetaPolicy_ANY,
											SubPolicy: config.AdminsPolicyKey,
										}),
									}),
								},
							},
						},
						"SampleConsortium2": {
							Groups: map[string]*cb.ConfigGroup{},
							Values: map[string]*cb.ConfigValue{
								config.ChannelCreationPolicyKey: {
									ModPolicy: "/Channel/Orderer/Admins",
									Value: marshalOrPanic(&cb.Policy{
										Type: 3,
										Value: marshalOrPanic(&cb.ImplicitMetaPolicy{
											Rule:      cb.ImplicitMetaPolicy_ANY,
											SubPolicy: config.AdminsPolicyKey,
										}),
									}),
								},
							},
						},
					},
					Values:   map[string]*cb.ConfigValue{},
					Policies: map[string]*cb.ConfigPolicy{},
				},
			},
		},
	}
}

// baseMSP creates a basic MSP struct for organization.
func baseMSP(t *testing.T) config.MSP {
	gt := NewGomegaWithT(t)

	certBlock, _ := pem.Decode([]byte(dummyCert))
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	gt.Expect(err).NotTo(HaveOccurred())

	privKeyBlock, _ := pem.Decode([]byte(dummyPrivateKey))
	privKey, err := x509.ParsePKCS8PrivateKey(privKeyBlock.Bytes)
	gt.Expect(err).NotTo(HaveOccurred())

	crlBlock, _ := pem.Decode([]byte(dummyCRL))
	crl, err := x509.ParseCRL(crlBlock.Bytes)
	gt.Expect(err).NotTo(HaveOccurred())

	return config.MSP{
		Name:              "MSPID",
		RootCerts:         []*x509.Certificate{cert},
		IntermediateCerts: []*x509.Certificate{cert},
		Admins:            []*x509.Certificate{cert},
		RevocationList:    []*pkix.CertificateList{crl},
		SigningIdentity: config.SigningIdentityInfo{
			PublicSigner: cert,
			PrivateSigner: config.KeyInfo{
				KeyIdentifier: "SKI-1",
				KeyMaterial:   privKey.(*ecdsa.PrivateKey),
			},
		},
		OrganizationalUnitIdentifiers: []config.OUIdentifier{
			{
				Certificate:                  cert,
				OrganizationalUnitIdentifier: "OUID",
			},
		},
		CryptoConfig: config.CryptoConfig{
			SignatureHashFamily:            "SHA3",
			IdentityIdentifierHashFunction: "SHA256",
		},
		TLSRootCerts:         []*x509.Certificate{cert},
		TLSIntermediateCerts: []*x509.Certificate{cert},
		NodeOus: config.NodeOUs{
			ClientOUIdentifier: config.OUIdentifier{
				Certificate:                  cert,
				OrganizationalUnitIdentifier: "OUID",
			},
			PeerOUIdentifier: config.OUIdentifier{
				Certificate:                  cert,
				OrganizationalUnitIdentifier: "OUID",
			},
			AdminOUIdentifier: config.OUIdentifier{
				Certificate:                  cert,
				OrganizationalUnitIdentifier: "OUID",
			},
			OrdererOUIdentifier: config.OUIdentifier{
				Certificate:                  cert,
				OrganizationalUnitIdentifier: "OUID",
			},
		},
	}
}
