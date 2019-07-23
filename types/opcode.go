package types

import "github.com/herdius/herdius-core/p2p/types/opcode"

// Opcode definitions
const (
	OpcodeChildBlockMessage           = opcode.Opcode(1111)
	OpcodeConnectionMessage           = opcode.Opcode(1112)
	OpcodeBlockHeightRequest          = opcode.Opcode(1113)
	OpcodeBlockResponse               = opcode.Opcode(1114)
	OpcodeAccountRequest              = opcode.Opcode(1115)
	OpcodeAccountResponse             = opcode.Opcode(1116)
	OpcodeTxRequest                   = opcode.Opcode(1117)
	OpcodeTxResponse                  = opcode.Opcode(1118)
	OpcodeTxDetailRequest             = opcode.Opcode(1119)
	OpcodeTxDetailResponse            = opcode.Opcode(1120)
	OpcodeTxsByAddressRequest         = opcode.Opcode(1121)
	OpcodeTxsResponse                 = opcode.Opcode(1122)
	OpcodeTxsByAssetAndAddressRequest = opcode.Opcode(1123)
	OpcodeTxUpdateRequest             = opcode.Opcode(1124)
	OpcodeTxUpdateResponse            = opcode.Opcode(1125)
	OpcodeTxDeleteRequest             = opcode.Opcode(1126)
	OpcodeTxLockedRequest             = opcode.Opcode(1127)
	OpcodeTxLockedResponse            = opcode.Opcode(1128)
	OpcodePing                        = opcode.Opcode(1129)
	OpcodePong                        = opcode.Opcode(1130)
	OpcodeTxRedeemRequest             = opcode.Opcode(1131)
	OpcodeTxRedeemResponse            = opcode.Opcode(1132)
	OpcodeTxsByBlockHeightRequest     = opcode.Opcode(1133)
	OpcodeLastBlockRequest            = opcode.Opcode(1134)
)
