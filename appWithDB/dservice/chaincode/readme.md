# install 

```sh

cd /tmp
wget https://www.multichain.com/download/multichain-1.0.6.tar.gz
tar -xvzf multichain-1.0.6.tar.gz
cd multichain-1.0.6
mv multichaind multichain-cli multichain-util /usr/local/bin

```


# start & connect


1. PC master



    ```sh
    export chainName='chain1'

    # create
    multichain-util create "$chainName"

    # edit config
    vi "~/.multichain/$chainName/params.dat"
    ## change default-network-port && default-network-port as follow
    ## default-network-port = 8347             # Default TCP/IP port for peer-to-peer connection with other nodes.
    ## default-rpc-port = 8346                 # Default TCP/IP port for incoming JSON-RPC API requests.

    # start
    multichaind "$chainName" -daemon

    ```

2. PC follower

    ```sh

    export chainName='chain1'
    export masterIP='...'
    export masterPort='8347'

    # connect master
    # 1. fetch chain1 configs
    # 2. create follower address
    # copy result from this command 
    multichaind "${chainName}@${masterIP}:${masterPort}"
    
    export followerAddress='1...'

    ```

3. PC master

    ```sh

    export followerAddress='1...'

    # grant follower
    multichain-cli "$chainName" grant "$followerAddress" connect,send,receive

    ```

4. PC follower

    ```sh

    # start
    multichaind "${chainName}" -daemon


    ```

> Now follower is connect to master and is able to send or receive asset


# Config BaaS


```sh

# fetch username & password
cat "~/.multichain/$chainName/multichain.conf"
# result likes
# rpcuser=multichainrpc
# rpcpassword=H2wiBAR7WFjtjsYGmBaKoTCByoUShpT7Fmtvsc7bNgk9

cat "~/.multichain/$chainName/params.dat" | grep 'default'
## result likes
## default-network-port = 8347             # Default TCP/IP port for peer-to-peer connection with other nodes.
## default-rpc-port = 8346                 # Default TCP/IP port for incoming JSON-RPC API requests.

```

```json
{
    "//": "put `rpc ip(127.0.0.1)` into key `btip`",
    "btip": "127.0.0.1",

    "//": "put `default-rpc-portinto` key `btport`",
    "btport": 8346,

    "//": "put `rpcuser` into key `btuser`",
    "btuser": "multichainrpc",

    "//": "put `rpcpassword` into key `btip`",
    "btpwd": "H2wiBAR7WFjtjsYGmBaKoTCByoUShpT7Fmtvsc7bNgk9",
}

```

> In real config.json, do not add K-V: { "//": "..." }


# cross

1. Bob prepare

    ```sh

    preparelockunspent '{}' true


    >>>>

    
    {
        "txid" : "b68d3d77bb665d43a2c1b45383c4c983d917e60a4b48ba656876e2018c5773bb",
        "vout" : 1
    }



    ```

2. Bob create asked 

    ```sh

    createrawexchange 'b68d3d77bb665d43a2c1b45383c4c983d917e60a4b48ba656876e2018c5773bb' 1 '{"asset1": 456}'

    0100000001bb73578c01e2766865ba484b0ae617d983c9c48353b4c1a2435d66bb773d8db6010000006b483045022100ec418db1d63358674dc4b0d3fda85c073055c86657e3ad17ad056a9d9de38d040220341833540278535740051b33502d3449aff9cbe3305f7fc27446f70f78ebb3f5832102400cd1eb3d7d9f257a3b458771e762067b4206687b47cdc66601f711af55c686ffffffff0100000000000000003776a914de3dd369db5ceb08c860b26db69d5a7d562e629988ac1c73706b71b7945a62f5220a35e562f4bd3230e35fc8010000000000007500000000

    ```

3. Bob issue 

    ```sh

    http://192.168.1.249:7000/cross/v0/issue

    {
        "fabric": {"address": "BobAddress", "count": 10.0},
        "bitcoin": null,
        "trade": 0
    }

    ```

4. Bob start

    ```sh

    http://192.168.1.249:7000/cross/v2/start

    {
        "asked": "0100000001bb73578c01e2766865ba484b0ae617d983c9c48353b4c1a2435d66bb773d8db6010000006b483045022100ec418db1d63358674dc4b0d3fda85c073055c86657e3ad17ad056a9d9de38d040220341833540278535740051b33502d3449aff9cbe3305f7fc27446f70f78ebb3f5832102400cd1eb3d7d9f257a3b458771e762067b4206687b47cdc66601f711af55c686ffffffff0100000000000000003776a914de3dd369db5ceb08c860b26db69d5a7d562e629988ac1c73706b71b7945a62f5220a35e562f4bd3230e35fc8010000000000007500000000",
        "address": "BobAddress",
        "offer": [{"name": "", "qty": 10}],
        "ask": [{"asset1": 456}],
        "channel": "chain1",
        "duration": 0
    }

    >>>

    {
        "success": true,
        "data": {
            "askedkey": "a71910542acaa33807fc6d397a78c091b74187338087a71c05cae46aebe9e3b04bf66bd087e21a7c4e4e10b3c211d0743cedb1761519bd5fdd80364b1ce4d345"
        }
    }

    ```

5. Alice Fetch

    ```sh

    http://192.168.1.249:7000/cross/v2/search
    {"askedkey":"a71910542acaa33807fc6d397a78c091b74187338087a71c05cae46aebe9e3b04bf66bd087e21a7c4e4e10b3c211d0743cedb1761519bd5fdd80364b1ce4d345"}

    >>>

    {
        "success": true,
        "txid": "8dc25816688bfb040e6f14d40aa9a36f2c6d8a5ffff5c4e7ee3b064be3d2e465",
        "data": {
            "askedaddress": "",
            "askedhex": "0100000001bb73578c01e2766865ba484b0ae617d983c9c48353b4c1a2435d66bb773d8db6010000006b483045022100ec418db1d63358674dc4b0d3fda85c073055c86657e3ad17ad056a9d9de38d040220341833540278535740051b33502d3449aff9cbe3305f7fc27446f70f78ebb3f5832102400cd1eb3d7d9f257a3b458771e762067b4206687b47cdc66601f711af55c686ffffffff0100000000000000003776a914de3dd369db5ceb08c860b26db69d5a7d562e629988ac1c73706b71b7945a62f5220a35e562f4bd3230e35fc8010000000000007500000000",
            "channel": "chain1",
            "completehex": "",
            "duration": 259200000000000,
            "finishedtxid": "",
            "offer": [
                {
                    "assetref": "",
                    "name": "",
                    "qty": 10
                }
            ],
            "offeraddress": "BobAddress",
            "start": "2018-07-24T07:23:41-04:00",
            "status": "waiting"
        }
    }

    ```

6. Alice prepare lock

    ```sh

    preparelockunspent '{"asset1": 456}' true

    >>>

    {
        "txid" : "62d247831acbcab46bd898eb3dc5a881c512ccdcc4e6f856b36cd2094078d691",
        "vout" : 0
    }


    ```

7. Alice balance the asked 

    ```sh

    completerawexchange 0100000001bb73578c01e2766865ba484b0ae617d983c9c48353b4c1a2435d66bb773d8db6010000006b483045022100ec418db1d63358674dc4b0d3fda85c073055c86657e3ad17ad056a9d9de38d040220341833540278535740051b33502d3449aff9cbe3305f7fc27446f70f78ebb3f5832102400cd1eb3d7d9f257a3b458771e762067b4206687b47cdc66601f711af55c686ffffffff0100000000000000003776a914de3dd369db5ceb08c860b26db69d5a7d562e629988ac1c73706b71b7945a62f5220a35e562f4bd3230e35fc8010000000000007500000000 62d247831acbcab46bd898eb3dc5a881c512ccdcc4e6f856b36cd2094078d691 0 '{}'

    >>>>

    0100000002bb73578c01e2766865ba484b0ae617d983c9c48353b4c1a2435d66bb773d8db6010000006b483045022100ec418db1d63358674dc4b0d3fda85c073055c86657e3ad17ad056a9d9de38d040220341833540278535740051b33502d3449aff9cbe3305f7fc27446f70f78ebb3f5832102400cd1eb3d7d9f257a3b458771e762067b4206687b47cdc66601f711af55c686ffffffff91d6784009d26cb356f8e6c4dccc12c581a8c53deb98d86bb4cacb1a8347d262000000006b4830450221008b24838b6b18ceec3d227ced3f2cfec759cbaefdae569b9563365f5bdd149f3a022039109c2092fa6f3e64889fa6a27f7784c9080120c0f7b537cc7da15f8caaed4f012102821ac29a07ea4373092408a9a4785919907ac9b8c93333b7d9814ba3239b04b1ffffffff0200000000000000003776a914de3dd369db5ceb08c860b26db69d5a7d562e629988ac1c73706b71b7945a62f5220a35e562f4bd3230e35fc8010000000000007500000000000000001976a9141a5170c1ab3381caf824e32d6bb059841960ae3a88ac00000000

    ```

8. Alice complate

    ```sh

    http://192.168.1.249:7000/cross/v2/complete

    {
        "askedkey": "a71910542acaa33807fc6d397a78c091b74187338087a71c05cae46aebe9e3b04bf66bd087e21a7c4e4e10b3c211d0743cedb1761519bd5fdd80364b1ce4d345",
        "completehex": "0100000002bb73578c01e2766865ba484b0ae617d983c9c48353b4c1a2435d66bb773d8db6010000006b483045022100ec418db1d63358674dc4b0d3fda85c073055c86657e3ad17ad056a9d9de38d040220341833540278535740051b33502d3449aff9cbe3305f7fc27446f70f78ebb3f5832102400cd1eb3d7d9f257a3b458771e762067b4206687b47cdc66601f711af55c686ffffffff91d6784009d26cb356f8e6c4dccc12c581a8c53deb98d86bb4cacb1a8347d262000000006b4830450221008b24838b6b18ceec3d227ced3f2cfec759cbaefdae569b9563365f5bdd149f3a022039109c2092fa6f3e64889fa6a27f7784c9080120c0f7b537cc7da15f8caaed4f012102821ac29a07ea4373092408a9a4785919907ac9b8c93333b7d9814ba3239b04b1ffffffff0200000000000000003776a914de3dd369db5ceb08c860b26db69d5a7d562e629988ac1c73706b71b7945a62f5220a35e562f4bd3230e35fc8010000000000007500000000000000001976a9141a5170c1ab3381caf824e32d6bb059841960ae3a88ac00000000",
        "address": "AliceAddress"
    }

    >>>

    {
        "success": true,
        "data": {
            "fxid": "b779fd490b0f22b02f94eee34578660dd4d4388032c95759d167b90390fe8483"
        }
    }

    ```

9. Check 

    ```sh

    http://192.168.1.249:7000/cross/v0/fetch

    {
        "fabric": {"address": "BobAddress", "count": 10.0},
        "bitcoin": {},
        "trade": 0
    }

    >>>


    {
        "success": true,
        "data": {
            "bitcoin": {
                "balances": [
                    {
                        "name": "asset1",
                        "assetref": "96-265-41467",
                        "qty": 544
                    }
                ]
            },
            "fabric": {
                "txid": "87f66eab7b8e2d69a2935a65c382669f32281740b4094c72284b6c451d2dc885",
                "account": {
                    "address": "BobAddress",
                    "count": 0
                }
            }
        }
    }

    ```