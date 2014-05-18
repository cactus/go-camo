import hashlib
import hmac
import base64


CAMO_HOST = 'https://img.example.com'


def camo_url(hmac_key, image_url):
    if image_url.startswith("https:"):
        return image_url
    b64digest = base64.urlsafe_b64encode(
        hmac.new(hmac_key, image_url, hashlib.sha1).digest()).strip('=')
    b64url = base64.urlsafe_b64encode(image_url).strip('=')
    requrl = '%s/%s/%s' % (CAMO_HOST, b64digest, b64url)
    return requrl


print camo_url("test", "http://golang.org/doc/gopher/frontpage.png")
# 'https://img.example.org/D23vHLFHsOhPOcvdxeoQyAJTpvM/aHR0cDovL2dvbGFuZy5vcmcvZG9jL2dvcGhlci9mcm9udHBhZ2UucG5n'

