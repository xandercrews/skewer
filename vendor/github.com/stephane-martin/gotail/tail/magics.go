package tail

const (
	sMAGICAAFS           = 0x5A3C69F0
	sMAGICACFS           = 0x61636673
	sMAGICADFS           = 0xADF5
	sMAGICAFFS           = 0xADFF
	sMAGICAFS            = 0x5346414F
	sMAGICANONINODEFS    = 0x09041934
	sMAGICAPFS           = 24
	sMAGICAUFS           = 0x61756673
	sMAGICAUTOFS         = 0x0187
	sMAGICBALLOONKVM     = 0x13661366
	sMAGICBEFS           = 0x42465331
	sMAGICBDEVFS         = 0x62646576
	sMAGICBFS            = 0x1BADFACE
	sMAGICBPFFS          = 0xCAFE4A11
	sMAGICBINFMTFS       = 0x42494E4D
	sMAGICBTRFS          = 0x9123683E
	sMAGICBTRFSTEST      = 0x73727279
	sMAGICCEPH           = 0x00C36400
	sMAGICCGROUP         = 0x0027E0EB
	sMAGICCGROUP2        = 0x63677270
	sMAGICCIFS           = 0xFF534D42
	sMAGICCODA           = 0x73757245
	sMAGICCOH            = 0x012FF7B7
	sMAGICCONFIGFS       = 0x62656570
	sMAGICCRAMFS         = 0x28CD3D45
	sMAGICCRAMFSWEND     = 0x453DCD28
	sMAGICDAXFS          = 0x64646178
	sMAGICDEBUGFS        = 0x64626720
	sMAGICDEVFS          = 0x1373
	sMAGICDEVPTS         = 0x1CD1
	sMAGICECRYPTFS       = 0xF15F
	sMAGICEFIVARFS       = 0xDE5E81E4
	sMAGICEFS            = 0x00414A53
	sMAGICEXOFS          = 0x5DF5
	sMAGICEXT            = 0x137D
	sMAGICEXT2           = 0xEF53
	sMAGICEXT2OLD        = 0xEF51
	sMAGICF2FS           = 0xF2F52010
	sMAGICFAT            = 0x4006
	sMAGICFHGFS          = 0x19830326
	sMAGICFUSEBLK        = 0x65735546
	sMAGICFUSECTL        = 0x65735543
	sMAGICFUTEXFS        = 0x0BAD1DEA
	sMAGICGFS            = 0x01161970
	sMAGICGPFS           = 0x47504653
	sMAGICHFS            = 0x4244
	sMAGICHFSPLUS        = 0x482B
	sMAGICHFSX           = 0x4858
	sMAGICHOSTFS         = 0x00C0FFEE
	sMAGICHPFS           = 0xF995E849
	sMAGICHUGETLBFS      = 0x958458F6
	sMAGICMTDINODEFS     = 0x11307854
	sMAGICIBRIX          = 0x013111A8
	sMAGICINOTIFYFS      = 0x2BAD1DEA
	sMAGICISOFS          = 0x9660
	sMAGICISOFSRWIN      = 0x4004
	sMAGICISOFSWIN       = 0x4000
	sMAGICJFFS           = 0x07C0
	sMAGICJFFS2          = 0x72B6
	sMAGICJFS            = 0x3153464A
	sMAGICKAFS           = 0x6B414653
	sMAGICLOGFS          = 0xC97E8168
	sMAGICLUSTRE         = 0x0BD00BD0
	sMAGICM1FS           = 0x5346314D
	sMAGICMINIX          = 0x137F
	sMAGICMINIX30        = 0x138F
	sMAGICMINIXV2        = 0x2468
	sMAGICMINIXV230      = 0x2478
	sMAGICMINIXV3        = 0x4D5A
	sMAGICMQUEUE         = 0x19800202
	sMAGICMSDOS          = 0x4D44
	sMAGICNCP            = 0x564C
	sMAGICNFS            = 0x6969
	sMAGICNFSD           = 0x6E667364
	sMAGICNILFS          = 0x3434
	sMAGICNSFS           = 0x6E736673
	sMAGICNTFS           = 0x5346544E
	sMAGICOPENPROM       = 0x9FA1
	sMAGICOCFS2          = 0x7461636F
	sMAGICOVERLAYFS      = 0x794C7630
	sMAGICPANFS          = 0xAAD7AAEA
	sMAGICPIPEFS         = 0x50495045
	sMAGICPRLFS          = 0x7C7C6673
	sMAGICPROC           = 0x9FA0
	sMAGICPSTOREFS       = 0x6165676C
	sMAGICQNX4           = 0x002F
	sMAGICQNX6           = 0x68191122
	sMAGICRAMFS          = 0x858458F6
	sMAGICRDTGROUP       = 0x07655821
	sMAGICREISERFS       = 0x52654973
	sMAGICROMFS          = 0x7275
	sMAGICRPCPIPEFS      = 0x67596969
	sMAGICSECURITYFS     = 0x73636673
	sMAGICSELINUX        = 0xF97CFF8C
	sMAGICSMACK          = 0x43415D53
	sMAGICSMB            = 0x517B
	sMAGICSMB2           = 0xFE534D42
	sMAGICSNFS           = 0xBEEFDEAD
	sMAGICSOCKFS         = 0x534F434B
	sMAGICSQUASHFS       = 0x73717368
	sMAGICSYSFS          = 0x62656572
	sMAGICSYSV2          = 0x012FF7B6
	sMAGICSYSV4          = 0x012FF7B5
	sMAGICTMPFS          = 0x01021994
	sMAGICTRACEFS        = 0x74726163
	sMAGICUBIFS          = 0x24051905
	sMAGICUDF            = 0x15013346
	sMAGICUFS            = 0x00011954
	sMAGICUFSBYTESWAPPED = 0x54190100
	sMAGICUSBDEVFS       = 0x9FA2
	sMAGICV9FS           = 0x01021997
	sMAGICVMHGFS         = 0xBACBACBC
	sMAGICVXFS           = 0xA501FCF5
	sMAGICVZFS           = 0x565A4653
	sMAGICWSLFS          = 0x53464846
	sMAGICXENFS          = 0xABBA1974
	sMAGICXENIX          = 0x012FF7B4
	sMAGICXFS            = 0x58465342
	sMAGICXIAFS          = 0x012FD16D
	sMAGICZFS            = 0x2FC12FC1
	sMAGICZSMALLOC       = 0x58295829
)
