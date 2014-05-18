import hashlib
import hmac


CAMO_HOST = 'https://img.example.com'


def camo_url(hmac_key, image_url):
    if image_url.startswith("https:"):
        return image_url
    hexdigest = hmac.new(hmac_key, image_url, hashlib.sha1).hexdigest()
    hexurl = image_url.encode('hex')
    requrl = '%s/%s/%s' % (CAMO_HOST, hexdigest, hexurl)
    return requrl


print camo_url("test", "http://golang.org/doc/gopher/frontpage.png")
# 'https://img.example.org/0f6def1cb147b0e84f39cbddc5ea10c80253a6f3/687474703a2f2f676f6c616e672e6f72672f646f632f676f706865722f66726f6e74706167652e706e67'
