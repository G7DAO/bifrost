# Setup Base bifrost and call

This checklist describes how to deploy a ERC20 token on Base and bridge it to Ethereum using the Optimism Canonical Bridge

## Deploy a ERC20 token on Base

```bash
bin/bifrost base optimism-mintable-erc-20-factory create-standard-l-2-token  \ 
   --keyfile $WB_WALLET \
   --rpc $BASE_SEPOLIA_RPC \
   --name-0 $NAME \
   --symbol $SYMBOL \
   --contract $OPTIMISM_ERC20_FACTORY_BASE_SEPOLIA \
   --remote-token $TOKEN_ON_ETHEREUM \
```

Output: Transaction Hash


## Bridge a ERC20 token from Base to Ethereum

```bash
bin/bifrost base l-1-standard-bridge bridge-erc-20 \
   --keyfile $WB_WALLET \
   --rpc $BASE_RPC \
   --contract $L1_STANDARD_BRIDGE_BASE \
   --local-token $TOKEN_ON_ETHEREUM \
   --remote-token $TOKEN_ON_BASE \
   --amount $AMOUNT \
   --min-gas-limit $MIN_GAS_LIMIT \
   --extra-data $EXTRA_DATA
```

Output: Transaction Hash