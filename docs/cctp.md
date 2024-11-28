# Setup CCTP bifrost and call

This checklist describes how to bridge tokens between any chain implementing CCTP

## Bifrost between any chain implementing CCTP

```bash
bin/bifrost cctp \ 
   --keyfile $WB_WALLET \
   --rpc $ETH_SEPOLIA_RPC \
   --recipient $RECIPIENT \
   --amount $AMOUNT \
   --token $TOKEN \
   --contract $CONTRACT \
   --domain $DOMAIN # 0 for ethereum, 3 for arbitrum
```

Output: Transaction Hash