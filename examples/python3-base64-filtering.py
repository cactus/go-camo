# Copyright (c) 2012-2023 Eli Janssen
# Use of this source code is governed by an MIT-style
# license that can be found in the LICENSE file.

# this example shows how filtering can be done on the url generation side.
# this example through https urls (no proxying required), and only allows http
# requests over port 80.

import hashlib
import hmac
import base64
from urllib.parse import urlsplit


CAMO_HOST = 'https://img.example.com'


def wrap_encode(data):
    """A little helper method to wrap b64encoding"""
    return base64.urlsafe_b64encode(data).strip(b'=').decode('utf-8')


def camo_url(hmac_key, image_url):
    url = urlsplit(image_url)

    if url.scheme == 'https':
        # pass through https, no need to proxy it to get security lock.
        # fast path. check this first.
        return image_url

    if url.scheme != 'http' or (':' in url.netloc and not url.netloc.endswith(':80')):
        # depending on application code, it may be more appropriate
        # to return a fixed url placeholder image of some kind (eg. 404 image url),
        # an empty string, or raise an exception that calling code handles.
        return "Nope!"

    hmac_key = hmac_key.encode() if isinstance(hmac_key, str) else hmac_key
    image_url = image_url.encode() if isinstance(image_url, str) else image_url

    b64digest = wrap_encode(
        hmac.new(hmac_key, image_url, hashlib.sha1).digest()
    )
    b64url = wrap_encode(image_url)
    requrl = '%s/%s/%s' % (CAMO_HOST, b64digest, b64url)
    return requrl


print(camo_url("test", "http://golang.org/doc/gopher/frontpage.png"))
# https://img.example.org/D23vHLFHsOhPOcvdxeoQyAJTpvM/aHR0cDovL2dvbGFuZy5vcmcvZG9jL2dvcGhlci9mcm9udHBhZ2UucG5n
print(camo_url("test", "http://golang.org:80/doc/gopher/frontpage.png"))
# https://img.example.com/8_b8SZkMlTYfsGFtkZS7SyJn37k/aHR0cDovL2dvbGFuZy5vcmc6ODAvZG9jL2dvcGhlci9mcm9udHBhZ2UucG5n
print(camo_url("test", "http://golang.org:8080/doc/gopher/frontpage.png"))
# Nope!
