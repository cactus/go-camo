# Copyright (c) 2012-2023 Eli Janssen
# Use of this source code is governed by an MIT-style
# license that can be found in the LICENSE file.

import hashlib
import hmac
import base64


CAMO_HOST = 'https://img.example.com'


def wrap_encode(data):
    """A little helper method to wrap b64encoding"""
    return base64.urlsafe_b64encode(data).strip(b'=').decode('utf-8')


def camo_url(hmac_key, image_url):
    if image_url.startswith("https:"):
        return image_url

    hmac_key = hmac_key.encode() if isinstance(hmac_key, str) else hmac_key
    image_url = image_url.encode() if isinstance(image_url, str) else image_url

    # setup the hmac construction
    mac = hmac.new(hmac_key, digestmod=hashlib.sha1)
    # add image_url
    mac.update(image_url)

    # generate digest
    digest = mac.digest()

    ## now build url
    b64digest = wrap_encode(digest)
    b64url = wrap_encode(image_url)
    requrl = '%s/%s/%s' % (CAMO_HOST, b64digest, b64url)
    return requrl


print(
    camo_url("test", "http://golang.org/doc/gopher/frontpage.png")
)
# https://img.example.org/D23vHLFHsOhPOcvdxeoQyAJTpvM/aHR0cDovL2dvbGFuZy5vcmcvZG9jL2dvcGhlci9mcm9udHBhZ2UucG5n