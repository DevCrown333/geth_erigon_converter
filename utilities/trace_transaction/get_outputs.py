from web3 import HTTPProvider
from web3 import Web3
import json

## This script will get you sample outputs for trace_block
## and its Geth equivalent debug_traceBlockByNumber


block_number = 16636490

block_hash = hex(block_number)

client = HTTPProvider('https://black-dark-meadow.quiknode.pro/1a15b493e4c3bc9e6aab158ea100180b05a1944a/')

erigon_output = client.make_request("trace_transaction", ["0x9e63085271890a141297039b3b711913699f1ee4db1acb667ad7ce304772036b"])

params = ["0x9e63085271890a141297039b3b711913699f1ee4db1acb667ad7ce304772036b",{ "tracer": "callTracer" }]

geth_output_call_tracer = client.make_request('debug_traceTransaction', params)

# Write both outputs to JSON files
with open('erigon_output.json', 'w') as f:
    json.dump(erigon_output, f)
    
with open('geth_output_call_tracer_version.json', 'w') as f:
    json.dump(geth_output_call_tracer, f)

# geth_params = [block_hash, {'tracer': 'prestateTracer'}]

# geth_output_prestate_tracer = client.make_request('debug_traceBlockByNumber', geth_params)

# with open('geth_output_prestate_tracer_version.json', 'w') as f:
#     json.dump(geth_output_prestate_tracer, f)
