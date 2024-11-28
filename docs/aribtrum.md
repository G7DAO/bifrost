# Setup arbitrum bifrost and call

This checklist describes how to bridge tokens

## Bifrost between arbitrum related chains (eth, arbitrum, orbit)

```bash
bin/bifrost arbitrum \ 
    --inbox $INBOX \
    --l1-rpc $L1_RPC \
    --l2-rpc $L2_RPC \
    --to $TO \
    --amount $AMOUNT \
    --l2-calldata $L2_CALLDATA \ 
    --keyfile $KEY \
    --password $PASSWORD 
```

Output: Transaction Hash