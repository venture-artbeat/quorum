#!/bin/sh
set -x
echo "Entrypoint script executed" 


geth init /data/genesis.json

# Update txnSizeLimit in the genesis.json file using sed
sed -i 's/"txnSizeLimit": 64/"txnSizeLimit": 512/' /data/genesis.json



if [ "$NODE_INDEX" == "1" ]; then
    cp -r /data/Node-0/* /root/.ethereum/geth
    cp /data/static-nodes.json /root/.ethereum/geth

elif [ "$NODE_INDEX" == "2" ] ; then
    cp -r /data/Node-1/* /root/.ethereum/geth
    cp /data/static-nodes.json /root/.ethereum/geth
fi


sed -i "s|<NODE1_IP>|$NODE1_IP|g" /root/.ethereum/geth/static-nodes.json
sed -i "s|<NODE2_IP>|$NODE2_IP|g" /root/.ethereum/geth/static-nodes.json

export PRIVATE_CONFIG=ignore
export ADDRESS=$(grep -o '"address": *"[^"]*"' ./root/.ethereum/geth/keystore/accountKeystore | grep -o '"[^"]*"$' | sed 's/"//g')

geth --mine  --miner.threads 1 --verbosity 5 --nodiscover --http --http.port 8546 --http.addr 0.0.0.0 --port 30303 --http.corsdomain "*" --http.vhosts "*" --nat "any" --http.api eth,web3,personal,net,miner --unlock ${ADDRESS} --password /root/.ethereum/geth/keystore/accountPassword --keystore ./root/.ethereum/geth/keystore --nodekey /root/.ethereum/geth/nodekey --ipcpath geth.ipc --allow-insecure-unlock 
