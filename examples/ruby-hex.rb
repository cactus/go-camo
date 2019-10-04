# Copyright (c) 2012-2019 Eli Janssen
# Use of this source code is governed by an MIT-style
# license that can be found in the LICENSE file.

require "openssl"

CAMO_HOST = "https://img.example.com"

def camo_url(hmac_key, image_url)
    if image_url.start_with?("https:")
        return image_url
    end
    hexdigest = OpenSSL::HMAC.hexdigest("sha1", hmac_key, image_url)
    hexurl = image_url.unpack("U*").collect{|x| x.to_s(16)}.join
    return "#{CAMO_HOST}/#{hexdigest}/#{hexurl}"
end

puts camo_url("test", "http://golang.org/doc/gopher/frontpage.png")
# 'https://img.example.org/0f6def1cb147b0e84f39cbddc5ea10c80253a6f3/687474703a2f2f676f6c616e672e6f72672f646f632f676f706865722f66726f6e74706167652e706e67'
