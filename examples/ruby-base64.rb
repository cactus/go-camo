# Copyright (c) 2012-2023 Eli Janssen
# Use of this source code is governed by an MIT-style
# license that can be found in the LICENSE file.

require "base64"
require "openssl"

CAMO_HOST = "https://img.example.com"

def camo_url(hmac_key, image_url)
    if image_url.start_with?("https:")
        return image_url
    end
    b64digest = Base64.urlsafe_encode64(OpenSSL::HMAC.digest("sha1", hmac_key, image_url)).delete("=")
    b64url = Base64.urlsafe_encode64(image_url).delete("=")
    return "#{CAMO_HOST}/#{b64digest}/#{b64url}"
end

puts camo_url("test", "http://golang.org/doc/gopher/frontpage.png")
# https://img.example.com/D23vHLFHsOhPOcvdxeoQyAJTpvM/aHR0cDovL2dvbGFuZy5vcmcvZG9jL2dvcGhlci9mcm9udHBhZ2UucG5n
