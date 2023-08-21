# Copyright (c) 2012-2019 Eli Janssen
# Use of this source code is governed by an MIT-style
# license that can be found in the LICENSE file.

import hashlib
import hmac
import base64
import json


CAMO_HOST = 'https://img.example.com'


def wrap_encode(data):
    """A little helper method to wrap b64encoding"""
    return base64.urlsafe_b64encode(data).strip(b'=').decode('utf-8')


def camo_url(hmac_key, image_url, extra_headers=None):
    if image_url.startswith("https:"):
        return image_url

    hmac_key = hmac_key.encode() if isinstance(hmac_key, str) else hmac_key
    image_url = image_url.encode() if isinstance(image_url, str) else image_url

    # setup the hmac construction
    mac = hmac.new(hmac_key, digestmod=hashlib.sha1)
    # add image_url
    mac.update(image_url)

    # if we have extra headers, encode them, and add to the hmac
    # this helps protect the header portion against tampering/modification
    if extra_headers:
        json_headers = json.dumps(extra_headers).encode('utf-8')
        # add json_headers to hmac, if present
        mac.update(json_headers)
        b64headers = wrap_encode(json_headers)
    else:
        b64headers = ""

    # generate digest
    digest = mac.digest()

    ## now build url
    b64digest = wrap_encode(digest)
    b64url = wrap_encode(image_url)
    requrl = '%s/%s/%s' % (CAMO_HOST, b64digest, b64url)
    # if we have extra headers, add it too
    if b64headers:
        requrl = requrl + f"/{b64headers}"
    return requrl


print(
    camo_url("test", "http://golang.org/doc/gopher/frontpage.png")
)
# https://img.example.org/D23vHLFHsOhPOcvdxeoQyAJTpvM/aHR0cDovL2dvbGFuZy5vcmcvZG9jL2dvcGhlci9mcm9udHBhZ2UucG5n

print(
    camo_url(
        "test",
        "http://golang.org/doc/gopher/frontpage.png",
        {"content-disposition": 'attachment; filename="image.png"'}
    )
)
# https://img.example.com/-hNoquWgyjNgzF7HXYyvGwteyLI/aHR0cDovL2dvbGFuZy5vcmcvZG9jL2dvcGhlci9mcm9udHBhZ2UucG5n/eyJjb250ZW50LWRpc3Bvc2l0aW9uIjogImF0dGFjaG1lbnQ7IGZpbGVuYW1lPVwiaW1hZ2UucG5nXCIifQ