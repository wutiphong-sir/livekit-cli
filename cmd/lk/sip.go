// Copyright 2023-2024 LiveKit, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/urfave/cli/v3"
)

//lint:file-ignore SA1019 we still support older APIs for compatibility

var (
	SIPCommands = []*cli.Command{
		{
			Name:  "sip",
			Usage: "Manage SIP Trunks, Dispatch Rules, and Participants",
			Commands: []*cli.Command{
				{
					Name:    "inbound",
					Aliases: []string{"in", "inbound-trunk"},
					Usage:   "Inbound SIP Trunk management",
					Commands: []*cli.Command{
						{
							Name:   "list",
							Usage:  "List all inbound SIP Trunks",
							Action: listSipInboundTrunk,
							Flags:  []cli.Flag{jsonFlag},
						},
						{
							Name:      "create",
							Usage:     "Create an inbound SIP Trunk",
							Action:    createSIPInboundTrunk,
							ArgsUsage: RequestDesc[livekit.CreateSIPInboundTrunkRequest](),
						},
						{
							Name:      "update",
							Usage:     "Update an inbound SIP Trunk",
							Action:    updateSIPInboundTrunk,
							ArgsUsage: RequestDesc[livekit.UpdateSIPInboundTrunkRequest](),
							Flags: []cli.Flag{
								&cli.StringFlag{
									Name:  "id",
									Usage: "ID for the trunk to update",
								},
								&cli.StringFlag{
									Name:  "name",
									Usage: "Sets a new name for the trunk",
								},
								&cli.StringSliceFlag{
									Name:  "numbers",
									Usage: "Sets a new list of numbers for the trunk",
								},
								&cli.StringFlag{
									Name:  "auth-user",
									Usage: "Set username for authentication",
								},
								&cli.StringFlag{
									Name:  "auth-pass",
									Usage: "Set password for authentication",
								},
							},
						},
						{
							Name:      "delete",
							Usage:     "Delete a SIP Trunk",
							Action:    deleteSIPTrunk,
							ArgsUsage: "SIPTrunk ID to delete",
						},
					},
				},
				{
					Name:    "outbound",
					Aliases: []string{"out", "outbound-trunk"},
					Usage:   "Outbound SIP Trunk management",
					Commands: []*cli.Command{
						{
							Name:   "list",
							Usage:  "List all outbound SIP Trunk",
							Action: listSipOutboundTrunk,
							Flags:  []cli.Flag{jsonFlag},
						},
						{
							Name:      "create",
							Usage:     "Create an outbound SIP Trunk",
							Action:    createSIPOutboundTrunk,
							ArgsUsage: RequestDesc[livekit.CreateSIPOutboundTrunkRequest](),
						},
						{
							Name:      "update",
							Usage:     "Update an outbound SIP Trunk",
							Action:    updateSIPOutboundTrunk,
							ArgsUsage: RequestDesc[livekit.UpdateSIPOutboundTrunkRequest](),
							Flags: []cli.Flag{
								&cli.StringFlag{
									Name:  "id",
									Usage: "ID for the trunk to update",
								},
								&cli.StringFlag{
									Name:  "name",
									Usage: "Sets a new name for the trunk",
								},
								&cli.StringFlag{
									Name:  "address",
									Usage: "Sets a new destination address for the trunk",
								},
								&cli.StringFlag{
									Name:  "transport",
									Usage: "Sets a new transport for the trunk",
								},
								&cli.StringSliceFlag{
									Name:  "numbers",
									Usage: "Sets a new list of numbers for the trunk",
								},
								&cli.StringFlag{
									Name:  "auth-user",
									Usage: "Set username for authentication",
								},
								&cli.StringFlag{
									Name:  "auth-pass",
									Usage: "Set password for authentication",
								},
							},
						},
						{
							Name:      "delete",
							Usage:     "Delete SIP Trunk",
							Action:    deleteSIPTrunk,
							ArgsUsage: "SIPTrunk ID to delete",
						},
					},
				},
				{
					Name:    "dispatch",
					Usage:   "SIP Dispatch Rule management",
					Aliases: []string{"dispatch-rule"},
					Commands: []*cli.Command{
						{
							Name:   "list",
							Usage:  "List all SIP Dispatch Rule",
							Action: listSipDispatchRule,
							Flags:  []cli.Flag{jsonFlag},
						},
						{
							Name:      "create",
							Usage:     "Create a SIP Dispatch Rule",
							Action:    createSIPDispatchRule,
							ArgsUsage: RequestDesc[livekit.CreateSIPDispatchRuleRequest](),
						},
						{
							Name:      "update",
							Usage:     "Update a SIP Dispatch Rule",
							Action:    updateSIPDispatchRule,
							ArgsUsage: RequestDesc[livekit.UpdateSIPDispatchRuleRequest](),
							Flags: []cli.Flag{
								&cli.StringFlag{
									Name:  "id",
									Usage: "ID for the rule to update",
								},
								&cli.StringFlag{
									Name:  "name",
									Usage: "Sets a new name for the rule",
								},
								&cli.StringSliceFlag{
									Name:  "trunks",
									Usage: "Sets a new list of trunk IDs",
								},
							},
						},
						{
							Name:      "delete",
							Usage:     "Delete SIP Dispatch Rule",
							Action:    deleteSIPDispatchRule,
							ArgsUsage: "SIPTrunk ID to delete",
						},
					},
				},
				{
					Name:  "participant",
					Usage: "SIP Participant management",
					Commands: []*cli.Command{
						{
							Name:      "create",
							Usage:     "Create a SIP Participant",
							Action:    createSIPParticipant,
							ArgsUsage: RequestDesc[livekit.CreateSIPParticipantRequest](),
							Flags: []cli.Flag{
								&cli.StringFlag{
									Name:  "trunk",
									Usage: "`SIP_TRUNK_ID` to use for the call (overrides json config)",
								},
								&cli.StringFlag{
									Name:  "number",
									Usage: "`SIP_NUMBER` to use for the call (overrides json config)",
								},
								&cli.StringFlag{
									Name:  "call",
									Usage: "`SIP_CALL_TO` number to use (overrides json config)",
								},
								&cli.StringFlag{
									Name:  "room",
									Usage: "`ROOM_NAME` to place the call to (overrides json config)",
								},
								&cli.BoolFlag{
									Name:  "wait",
									Usage: "wait for the call to dial (overrides json config)",
								},
								&cli.DurationFlag{
									Name:  "timeout",
									Usage: "timeout for the call to dial (requires wait flag)",
									Value: 80 * time.Second,
								},
							},
						},
						{
							Name:   "transfer",
							Usage:  "Transfer a SIP Participant",
							Action: transferSIPParticipant,
							Flags: []cli.Flag{
								roomFlag,
								identityFlag,
								&cli.StringFlag{
									Name:     "to",
									Required: true,
									Usage:    "`SIP URL` to transfer the call to. Use 'tel:<phone number>' to transfer to a phone",
								},
								&cli.BoolFlag{
									Name:  "play-dialtone",
									Usage: "if set, a dial tone will be played to the SIP participant while the transfer is being attempted",
								},
							},
						},
					},
				},
			},
		},

		// Deprecated commands kept for compatibility
		{
			Hidden: true, // deprecated: use `sip trunk list`
			Name:   "list-sip-trunk",
			Usage:  "List all SIP trunk",
			Action: listSipTrunk,
		},
		{
			Hidden: true, // deprecated: use `sip trunk delete`
			Name:   "delete-sip-trunk",
			Usage:  "Delete SIP Trunk",
			Action: deleteSIPTrunkLegacy,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "id",
					Usage:    "SIPTrunk ID",
					Required: true,
				},
			},
		},
		{
			Hidden: true, // deprecated: use `sip dispatch create`
			Name:   "create-sip-dispatch-rule",
			Usage:  "Create a SIP Dispatch Rule",
			Action: createSIPDispatchRuleLegacy,
			Flags: []cli.Flag{
				RequestFlag[livekit.CreateSIPDispatchRuleRequest](),
			},
		},
		{
			Hidden: true, // deprecated: use `sip dispatch list`
			Name:   "list-sip-dispatch-rule",
			Usage:  "List all SIP Dispatch Rule",
			Action: listSipDispatchRule,
		},
		{
			Hidden: true, // deprecated: use `sip dispatch delete`
			Name:   "delete-sip-dispatch-rule",
			Usage:  "Delete SIP Dispatch Rule",
			Action: deleteSIPDispatchRuleLegacy,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "id",
					Usage:    "SIPDispatchRule ID",
					Required: true,
				},
			},
		},
		{
			Hidden: true, // deprecated: use `sip participant create`
			Name:   "create-sip-participant",
			Usage:  "Create a SIP Participant",
			Action: createSIPParticipantLegacy,
			Flags: []cli.Flag{
				RequestFlag[livekit.CreateSIPParticipantRequest](),
			},
		},
	}
)

func listUpdateFlag(cmd *cli.Command, setName string) *livekit.ListUpdate {
	if !cmd.IsSet(setName) {
		return nil
	}
	val := cmd.StringSlice(setName)
	if len(val) == 1 && val[0] == "" {
		val = []string{}
	}
	return &livekit.ListUpdate{Set: val}
}

func createSIPClient(cmd *cli.Command) (*lksdk.SIPClient, error) {
	pc, err := loadProjectDetails(cmd)
	if err != nil {
		return nil, err
	}
	return lksdk.NewSIPClient(pc.URL, pc.APIKey, pc.APISecret, withDefaultClientOpts(pc)...), nil
}

func createSIPInboundTrunk(ctx context.Context, cmd *cli.Command) error {
	cli, err := createSIPClient(cmd)
	if err != nil {
		return err
	}
	return createAndPrintReqs(ctx, cmd, nil, cli.CreateSIPInboundTrunk, printSIPInboundTrunkID)
}

func updateSIPInboundTrunk(ctx context.Context, cmd *cli.Command) error {
	cli, err := createSIPClient(cmd)
	if err != nil {
		return err
	}
	id := cmd.String("id")
	if cmd.Args().Len() > 1 {
		return errors.New("expected one JSON file or flags")
	}
	if cmd.Args().Len() == 1 {
		// Update from the JSON
		req, err := ReadRequestFileOrLiteral[livekit.SIPInboundTrunkInfo](cmd.Args().First())
		if err != nil {
			return fmt.Errorf("could not read request: %w", err)
		}
		if id == "" {
			id = req.SipTrunkId
		}
		req.SipTrunkId = ""
		if id == "" {
			return errors.New("no ID specified, use flag or set it in JSON")
		}
		info, err := cli.UpdateSIPInboundTrunk(ctx, &livekit.UpdateSIPInboundTrunkRequest{
			SipTrunkId: id,
			Action: &livekit.UpdateSIPInboundTrunkRequest_Replace{
				Replace: req,
			},
		})
		if err != nil {
			return err
		}
		printSIPInboundTrunkID(info)
		return err
	}
	// Update from flags
	if id == "" {
		return errors.New("no ID specified")
	}
	req := &livekit.SIPInboundTrunkUpdate{}
	if val := cmd.String("name"); val != "" {
		req.Name = &val
	}
	if val := cmd.String("auth-user"); val != "" {
		req.AuthUsername = &val
	}
	if val := cmd.String("auth-pass"); val != "" {
		req.AuthPassword = &val
	}
	req.Numbers = listUpdateFlag(cmd, "numbers")
	info, err := cli.UpdateSIPInboundTrunk(ctx, &livekit.UpdateSIPInboundTrunkRequest{
		SipTrunkId: id,
		Action: &livekit.UpdateSIPInboundTrunkRequest_Update{
			Update: req,
		},
	})
	if err != nil {
		return err
	}
	printSIPInboundTrunkID(info)
	return err
}

func createSIPOutboundTrunk(ctx context.Context, cmd *cli.Command) error {
	cli, err := createSIPClient(cmd)
	if err != nil {
		return err
	}
	return createAndPrintReqs(ctx, cmd, nil, cli.CreateSIPOutboundTrunk, printSIPOutboundTrunkID)
}

func updateSIPOutboundTrunk(ctx context.Context, cmd *cli.Command) error {
	cli, err := createSIPClient(cmd)
	if err != nil {
		return err
	}
	id := cmd.String("id")
	if cmd.Args().Len() > 1 {
		return errors.New("expected one JSON file or flags")
	}
	if cmd.Args().Len() == 1 {
		// Update from the JSON
		req, err := ReadRequestFileOrLiteral[livekit.SIPOutboundTrunkInfo](cmd.Args().First())
		if err != nil {
			return fmt.Errorf("could not read request: %w", err)
		}
		if id == "" {
			id = req.SipTrunkId
		}
		req.SipTrunkId = ""
		if id == "" {
			return errors.New("no ID specified, use flag or set it in JSON")
		}
		info, err := cli.UpdateSIPOutboundTrunk(ctx, &livekit.UpdateSIPOutboundTrunkRequest{
			SipTrunkId: id,
			Action: &livekit.UpdateSIPOutboundTrunkRequest_Replace{
				Replace: req,
			},
		})
		if err != nil {
			return err
		}
		printSIPOutboundTrunkID(info)
		return err
	}
	// Update from flags
	if id == "" {
		return errors.New("no ID specified")
	}
	req := &livekit.SIPOutboundTrunkUpdate{}
	if val := cmd.String("name"); val != "" {
		req.Name = &val
	}
	if val := cmd.String("address"); val != "" {
		req.Address = &val
	}
	if val := cmd.String("transport"); val != "" {
		val = strings.ToUpper(val)
		if !strings.HasPrefix(val, "SIP_TRANSPORT_") {
			val = "SIP_TRANSPORT_" + val
		}
		trv, ok := livekit.SIPTransport_value[val]
		if !ok {
			return fmt.Errorf("unsupported transport: %q", val)
		}
		tr := livekit.SIPTransport(trv)
		req.Transport = &tr
	}
	if val := cmd.String("auth-user"); val != "" {
		req.AuthUsername = &val
	}
	if val := cmd.String("auth-pass"); val != "" {
		req.AuthPassword = &val
	}
	req.Numbers = listUpdateFlag(cmd, "numbers")
	info, err := cli.UpdateSIPOutboundTrunk(ctx, &livekit.UpdateSIPOutboundTrunkRequest{
		SipTrunkId: id,
		Action: &livekit.UpdateSIPOutboundTrunkRequest_Update{
			Update: req,
		},
	})
	if err != nil {
		return err
	}
	printSIPOutboundTrunkID(info)
	return err
}

func userPass(user string, hasPass bool) string {
	if user == "" && !hasPass {
		return ""
	}
	passStr := ""
	if hasPass {
		passStr = "****"
	}
	return user + " / " + passStr
}

func printHeaders(m map[string]string) string {
	if len(m) == 0 {
		return ""
	}
	keys := slices.Collect(maps.Keys(m))
	slices.Sort(keys)
	var buf strings.Builder
	for i, key := range keys {
		if i != 0 {
			buf.WriteString("\n")
		}
		v := m[key]
		buf.WriteString(key)
		buf.WriteString("=")
		buf.WriteString(v)
	}
	return buf.String()
}

func printHeaderMaps(arr ...map[string]string) string {
	var out []string
	for _, m := range arr {
		s := printHeaders(m)
		if s == "" {
			continue
		}
		out = append(out, s)
	}
	if len(out) == 0 {
		return ""
	}
	return strings.Join(out, "\n\n")
}

func listSipTrunk(ctx context.Context, cmd *cli.Command) error {
	cli, err := createSIPClient(cmd)
	if err != nil {
		return err
	}
	//lint:ignore SA1019 we still support it
	return listAndPrint(ctx, cmd, cli.ListSIPTrunk, &livekit.ListSIPTrunkRequest{}, []string{
		"SipTrunkID", "Name", "Kind", "Number",
		"AllowAddresses", "AllowNumbers", "InboundAuth",
		"OutboundAddress", "OutboundAuth",
		"Metadata",
	}, func(item *livekit.SIPTrunkInfo) []string {
		inboundNumbers := item.InboundNumbers
		for _, re := range item.InboundNumbersRegex {
			inboundNumbers = append(inboundNumbers, "regexp("+re+")")
		}
		return []string{
			item.SipTrunkId, item.Name, strings.TrimPrefix(item.Kind.String(), "TRUNK_"), item.OutboundNumber,
			strings.Join(item.InboundAddresses, ","), strings.Join(inboundNumbers, ","), userPass(item.InboundUsername, item.InboundPassword != ""),
			item.OutboundAddress, userPass(item.OutboundUsername, item.OutboundPassword != ""),
			item.Metadata,
		}
	})
}

func listSipInboundTrunk(ctx context.Context, cmd *cli.Command) error {
	cli, err := createSIPClient(cmd)
	if err != nil {
		return err
	}
	return listAndPrint(ctx, cmd, cli.ListSIPInboundTrunk, &livekit.ListSIPInboundTrunkRequest{}, []string{
		"SipTrunkID", "Name", "Numbers",
		"AllowedAddresses", "AllowedNumbers",
		"Authentication",
		"Encryption",
		"Headers",
		"Metadata",
	}, func(item *livekit.SIPInboundTrunkInfo) []string {
		return []string{
			item.SipTrunkId, item.Name, strings.Join(item.Numbers, ","),
			strings.Join(item.AllowedAddresses, ","), strings.Join(item.AllowedNumbers, ","),
			userPass(item.AuthUsername, item.AuthPassword != ""),
			strings.TrimPrefix(item.MediaEncryption.String(), "SIP_MEDIA_ENCRYPT_"),
			printHeaderMaps(item.Headers, item.HeadersToAttributes),
			item.Metadata,
		}
	})
}

func listSipOutboundTrunk(ctx context.Context, cmd *cli.Command) error {
	cli, err := createSIPClient(cmd)
	if err != nil {
		return err
	}
	return listAndPrint(ctx, cmd, cli.ListSIPOutboundTrunk, &livekit.ListSIPOutboundTrunkRequest{}, []string{
		"SipTrunkID", "Name",
		"Address", "Transport",
		"Numbers",
		"Authentication",
		"Encryption",
		"Headers",
		"Metadata",
	}, func(item *livekit.SIPOutboundTrunkInfo) []string {
		return []string{
			item.SipTrunkId, item.Name,
			item.Address, strings.TrimPrefix(item.Transport.String(), "SIP_TRANSPORT_"),
			strings.Join(item.Numbers, ","),
			userPass(item.AuthUsername, item.AuthPassword != ""),
			strings.TrimPrefix(item.MediaEncryption.String(), "SIP_MEDIA_ENCRYPT_"),
			printHeaderMaps(item.Headers, item.HeadersToAttributes),
			item.Metadata,
		}
	})
}

func deleteSIPTrunk(ctx context.Context, cmd *cli.Command) error {
	cli, err := createSIPClient(cmd)
	if err != nil {
		return err
	}
	return forEachID(ctx, cmd, func(ctx context.Context, id string) error {
		info, err := cli.DeleteSIPTrunk(ctx, &livekit.DeleteSIPTrunkRequest{
			SipTrunkId: id,
		})
		if err != nil {
			return err
		}
		printSIPTrunkID(info)
		return nil
	})
}

func deleteSIPTrunkLegacy(ctx context.Context, cmd *cli.Command) error {
	cli, err := createSIPClient(cmd)
	if err != nil {
		return err
	}
	info, err := cli.DeleteSIPTrunk(ctx, &livekit.DeleteSIPTrunkRequest{
		SipTrunkId: cmd.String("id"),
	})
	if err != nil {
		return err
	}
	printSIPTrunkID(info)
	return nil
}

func printSIPTrunkID(info *livekit.SIPTrunkInfo) {
	fmt.Printf("SIPTrunkID: %v\n", info.GetSipTrunkId())
}

func printSIPInboundTrunkID(info *livekit.SIPInboundTrunkInfo) {
	fmt.Printf("SIPTrunkID: %v\n", info.GetSipTrunkId())
}

func printSIPOutboundTrunkID(info *livekit.SIPOutboundTrunkInfo) {
	fmt.Printf("SIPTrunkID: %v\n", info.GetSipTrunkId())
}

func createSIPDispatchRule(ctx context.Context, cmd *cli.Command) error {
	cli, err := createSIPClient(cmd)
	if err != nil {
		return err
	}
	return createAndPrintReqs(ctx, cmd, nil, cli.CreateSIPDispatchRule, printSIPDispatchRuleID)
}

func createSIPDispatchRuleLegacy(ctx context.Context, cmd *cli.Command) error {
	cli, err := createSIPClient(cmd)
	if err != nil {
		return err
	}
	return createAndPrintLegacy(ctx, cmd, cli.CreateSIPDispatchRule, printSIPDispatchRuleID)
}

func updateSIPDispatchRule(ctx context.Context, cmd *cli.Command) error {
	cli, err := createSIPClient(cmd)
	if err != nil {
		return err
	}
	id := cmd.String("id")
	if cmd.Args().Len() > 1 {
		return errors.New("expected one JSON file or flags")
	}
	if cmd.Args().Len() == 1 {
		// Update from the JSON
		req, err := ReadRequestFileOrLiteral[livekit.SIPDispatchRuleInfo](cmd.Args().First())
		if err != nil {
			return fmt.Errorf("could not read request: %w", err)
		}
		if id == "" {
			id = req.SipDispatchRuleId
		}
		req.SipDispatchRuleId = ""
		if id == "" {
			return errors.New("no ID specified, use flag or set it in JSON")
		}
		info, err := cli.UpdateSIPDispatchRule(ctx, &livekit.UpdateSIPDispatchRuleRequest{
			SipDispatchRuleId: id,
			Action: &livekit.UpdateSIPDispatchRuleRequest_Replace{
				Replace: req,
			},
		})
		if err != nil {
			return err
		}
		printSIPDispatchRuleID(info)
		return err
	}
	// Update from flags
	if id == "" {
		return errors.New("no ID specified")
	}
	req := &livekit.SIPDispatchRuleUpdate{}
	if val := cmd.String("name"); val != "" {
		req.Name = &val
	}
	req.TrunkIds = listUpdateFlag(cmd, "trunks")
	info, err := cli.UpdateSIPDispatchRule(ctx, &livekit.UpdateSIPDispatchRuleRequest{
		SipDispatchRuleId: id,
		Action: &livekit.UpdateSIPDispatchRuleRequest_Update{
			Update: req,
		},
	})
	if err != nil {
		return err
	}
	printSIPDispatchRuleID(info)
	return err
}

func listSipDispatchRule(ctx context.Context, cmd *cli.Command) error {
	cli, err := createSIPClient(cmd)
	if err != nil {
		return err
	}
	return listAndPrint(ctx, cmd, cli.ListSIPDispatchRule, &livekit.ListSIPDispatchRuleRequest{}, []string{
		"SipDispatchRuleID", "Name", "SipTrunks", "Type", "RoomName", "Pin",
		"Attributes", "Agents",
	}, func(item *livekit.SIPDispatchRuleInfo) []string {
		var room, typ, pin string
		switch r := item.GetRule().GetRule().(type) {
		case *livekit.SIPDispatchRule_DispatchRuleDirect:
			room = r.DispatchRuleDirect.RoomName
			pin = r.DispatchRuleDirect.Pin
			typ = "Direct"
		case *livekit.SIPDispatchRule_DispatchRuleIndividual:
			room = r.DispatchRuleIndividual.RoomPrefix + "_<caller>_<random>"
			pin = r.DispatchRuleIndividual.Pin
			typ = "Individual (Caller)"
		case *livekit.SIPDispatchRule_DispatchRuleCallee:
			room = r.DispatchRuleCallee.RoomPrefix + "<callee>"
			if r.DispatchRuleCallee.Randomize {
				room += "_<random>"
			}
			pin = r.DispatchRuleCallee.Pin
			typ = "Callee"
		}
		trunks := strings.Join(item.TrunkIds, ",")
		if trunks == "" {
			trunks = "<any>"
		}
		var agents []string
		if item.RoomConfig != nil {
			for _, agent := range item.RoomConfig.Agents {
				agents = append(agents, agent.AgentName)
			}
		}
		return []string{
			item.SipDispatchRuleId, item.Name, trunks, typ, room, pin,
			fmt.Sprintf("%v", item.Attributes), strings.Join(agents, ","),
		}
	})
}

func deleteSIPDispatchRule(ctx context.Context, cmd *cli.Command) error {
	cli, err := createSIPClient(cmd)
	if err != nil {
		return err
	}
	return forEachID(ctx, cmd, func(ctx context.Context, id string) error {
		info, err := cli.DeleteSIPDispatchRule(ctx, &livekit.DeleteSIPDispatchRuleRequest{
			SipDispatchRuleId: id,
		})
		if err != nil {
			return err
		}
		printSIPDispatchRuleID(info)
		return nil
	})
}

func deleteSIPDispatchRuleLegacy(ctx context.Context, cmd *cli.Command) error {
	cli, err := createSIPClient(cmd)
	if err != nil {
		return err
	}
	info, err := cli.DeleteSIPDispatchRule(ctx, &livekit.DeleteSIPDispatchRuleRequest{
		SipDispatchRuleId: cmd.String("id"),
	})
	if err != nil {
		return err
	}
	printSIPDispatchRuleID(info)
	return nil
}

func printSIPDispatchRuleID(info *livekit.SIPDispatchRuleInfo) {
	fmt.Printf("SIPDispatchRuleID: %v\n", info.SipDispatchRuleId)
}

func createSIPParticipant(ctx context.Context, cmd *cli.Command) error {
	cli, err := createSIPClient(cmd)
	if err != nil {
		return err
	}
	return createAndPrintReqs(ctx, cmd, func(req *livekit.CreateSIPParticipantRequest) error {
		if v := cmd.String("trunk"); v != "" {
			req.SipTrunkId = v
		}
		if v := cmd.String("number"); v != "" {
			req.SipNumber = v
		}
		if v := cmd.String("call"); v != "" {
			req.SipCallTo = v
		}
		if v := cmd.String("room"); v != "" {
			req.RoomName = v
		}
		if cmd.Bool("wait") {
			req.WaitUntilAnswered = true
		}
		return req.Validate()
	}, func(ctx context.Context, req *livekit.CreateSIPParticipantRequest) (*livekit.SIPParticipantInfo, error) {
		// CreateSIPParticipant will wait for LiveKit Participant to be created and that can take some time.
		// Default deadline is too short, thus, we must set a higher deadline for it.
		timeout := 30 * time.Second
		if req.WaitUntilAnswered {
			if dt := cmd.Duration("timeout"); dt != 0 {
				timeout = dt
			}
		}
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		resp, err := cli.CreateSIPParticipant(ctx, req)
		if e := lksdk.SIPStatusFrom(err); e != nil {
			msg := e.Status
			if msg == "" {
				msg = e.Code.ShortName()
			}
			fmt.Printf("SIPStatusCode: %d\n", e.Code)
			fmt.Printf("SIPStatus: %s\n", msg)
		}
		return resp, err
	}, printSIPParticipantInfo)
}

func createSIPParticipantLegacy(ctx context.Context, cmd *cli.Command) error {
	cli, err := createSIPClient(cmd)
	if err != nil {
		return err
	}
	return createAndPrintLegacy(ctx, cmd, func(ctx context.Context, req *livekit.CreateSIPParticipantRequest) (*livekit.SIPParticipantInfo, error) {
		// CreateSIPParticipant will wait for LiveKit Participant to be created and that can take some time.
		// Default deadline is too short, thus, we must set a higher deadline for it.
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		return cli.CreateSIPParticipant(ctx, req)
	}, printSIPParticipantInfo)
}

func transferSIPParticipant(ctx context.Context, cmd *cli.Command) error {
	roomName, identity := participantInfoFromArgOrFlags(cmd)
	to := cmd.String("to")
	dialtone := cmd.Bool("play-dialtone")

	req := livekit.TransferSIPParticipantRequest{
		RoomName:            roomName,
		ParticipantIdentity: identity,
		TransferTo:          to,
		PlayDialtone:        dialtone,
	}

	cli, err := createSIPClient(cmd)
	if err != nil {
		return err
	}

	_, err = cli.TransferSIPParticipant(ctx, &req)
	if err != nil {
		return err
	}

	return nil
}

func printSIPParticipantInfo(info *livekit.SIPParticipantInfo) {
	fmt.Printf("SIPCallID: %v\n", info.SipCallId)
	fmt.Printf("ParticipantID: %v\n", info.ParticipantId)
	fmt.Printf("ParticipantIdentity: %v\n", info.ParticipantIdentity)
	fmt.Printf("RoomName: %v\n", info.RoomName)
}
