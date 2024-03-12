#!/bin/sh
set -x
echo "Entrypoint script executed" 


geth init /data/genesis.json

# Update txnSizeLimit in the genesis.json file using sed
sed -i 's/"txnSizeLimit": 64/"txnSizeLimit": 512/' /data/genesis.json



if [ "$NODE_INDEX" == "1" ]; then
    cp -r /data/Node-1/* /root/.ethereum/geth
    cp /data/static-nodes.json /root/.ethereum/geth
elif [ "$NODE_INDEX" == "2" ] ; then
    cp -r /data/Node-2/* /root/.ethereum/geth
    cp /data/static-nodes.json /root/.ethereum/geth
elif [ "$NODE_INDEX" == "3" ]; then  
   cp -r /data/Node-3/* /root/.ethereum/geth
   cp /data/static-nodes.json /root/.ethereum/geth
elif [ "$NODE_INDEX" == "4" ]; then  
   cp -r /data/Node-4/* /root/.ethereum/geth
   cp /data/static-nodes.json /root/.ethereum/geth
elif [ "$NODE_INDEX" == "5" ]; then  
   cp -r /data/Node-5/* /root/.ethereum/geth
   cp /data/static-nodes.json /root/.ethereum/geth
fi


sed -i "s|<NODE1_IP>|$NODE1_IP|g" /root/.ethereum/geth/static-nodes.json
sed -i "s|<NODE2_IP>|$NODE2_IP|g" /root/.ethereum/geth/static-nodes.json
sed -i "s|<NODE3_IP>|$NODE3_IP|g" /root/.ethereum/geth/static-nodes.json
sed -i "s|<NODE4_IP>|$NODE4_IP|g" /root/.ethereum/geth/static-nodes.json
sed -i "s|<NODE5_IP>|$NODE5_IP|g" /root/.ethereum/geth/static-nodes.json

export PRIVATE_CONFIG=ignore
export ADDRESS=$(grep -o '"address": *"[^"]*"' ./root/.ethereum/geth/keystore/accountKeystore | grep -o '"[^"]*"$' | sed 's/"//g')

geth --mine  --miner.threads 1 --verbosity 3 --nodiscover \
     --http --http.api admin,debug,eth,net,web3 --http.port 8545 --http.addr 0.0.0.0 --port 30303 --http.corsdomain "*" --http.vhosts "*" \
     --ws --ws.api admin,eth,debug,web3,personal,net,miner --ws.addr 0.0.0.0 --ws.port 8546 --ws.origins '*' --ws.api eth,net,web3 \
     --unlock ${ADDRESS} --password /root/.ethereum/geth/keystore/accountPassword \
     --keystore ./root/.ethereum/geth/keystore --nodekey /root/.ethereum/geth/nodekey \
     --ipcpath geth.ipc --allow-insecure-unlock 
