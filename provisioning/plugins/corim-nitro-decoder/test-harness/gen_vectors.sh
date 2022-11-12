# Copyright 2022 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

cocli comid create --template AWSNitroComid.json
cocli corim create -m AWSNitroComid.cbor -t corimMini.json -o unsigned_corim.cbor
echo 'var unsignedCorim = `'
xxd -p unsigned_corim.cbor
echo '`'

cocli comid create --template AWSNitroComidDualKey.json
cocli corim create -m AWSNitroComidDualKey.cbor -t corimMini.json -o unsigned_corim_dual_key.cbor
echo 'var unsignedCorimDualKey = `'
xxd -p unsigned_corim_dual_key.cbor
echo '`'

cocli comid create --template AWSNitroComidNoImplId.json
cocli corim create -m AWSNitroComidNoImplId.cbor -t corimMini.json -o unsigned_corim_no_impl_id.cbor
echo 'var unsignedCorimNoImplId = `'
xxd -p unsigned_corim_no_impl_id.cbor
echo '`'
