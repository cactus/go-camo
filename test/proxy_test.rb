# Copyright (c) 2010 Corey Donohoe, Rick Olson
#
# Permission is hereby granted, free of charge, to any person obtaining
# a copy of this software and associated documentation files (the
# "Software"), to deal in the Software without restriction, including
# without limitation the rights to use, copy, modify, merge, publish,
# distribute, sublicense, and/or sell copies of the Software, and to
# permit persons to whom the Software is furnished to do so, subject to
# the following conditions:
#
# The above copyright notice and this permission notice shall be
# included in all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
# EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
# MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
# NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
# LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
# OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
# WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.


# This file is originally from the camo project.
#   https://github.com/atmos/camo
# Modified to remove Query String tests (CamoProxyQueryStringTest).

require 'rubygems'
require 'base64'
require 'openssl'
require 'rest_client'
require 'addressable/uri'
require 'contest'

class CamoProxyTests < Test::Unit::TestCase
  def config
      { 'key'  => ENV['CAMO_KEY']  || "0x24FEEDFACEDEADBEEFCAFE",
        'host' => ENV['CAMO_HOST'] || "http://localhost:8080" }
  end

  def hexenc(image_url)
    image_url.to_enum(:each_byte).map { |byte| "%02x" % byte }.join
  end

  def request_uri(image_url)
    hexdigest = OpenSSL::HMAC.hexdigest(
      OpenSSL::Digest::Digest.new('sha1'), config['key'], image_url)
    encoded_image_url = hexenc(image_url)
    "#{config['host']}/#{hexdigest}/#{encoded_image_url}"
  end

  def request(image_url)
    RestClient.get(request_uri(image_url))
  end

  describe "endpoints requiring camoproxy" do
    should "proxy valid image url" do
      response = request('http://media.ebaumsworld.com/picture/Mincemeat/Pimp.jpg')
      assert_equal(200, response.code)
    end

    should "proxy valid image url with crazy subdomain" do
      response = request('http://27.media.tumblr.com/tumblr_lkp6rdDfRi1qce6mto1_500.jpg')
      assert_equal(200, response.code)
    end

    should "proxy valid google chart url" do
      response = request('http://chart.apis.google.com/chart?chs=920x200&chxl=0:%7C2010-08-13%7C2010-09-12%7C2010-10-12%7C2010-11-11%7C1:%7C0%7C0%7C0%7C0%7C0%7C0&chm=B,EBF5FB,0,0,0&chco=008Cd6&chls=3,1,0&chg=8.3,20,1,4&chd=s:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA&chxt=x,y&cht=lc')
      assert_equal(200, response.code)
    end

    should "proxy valid chunked image file" do
      response = request('http://www.igvita.com/posts/12/spdyproxy-diagram.png')
      assert_equal(200, response.code)
      assert_nil(response.headers[:content_length])
    end

    should "follow redirects" do
      response = request('http://cl.ly/1K0X2Y2F1P0o3z140p0d/boom-headshot.gif')
      assert_equal(200, response.code)
    end

    should "follow redirects formatted strangely" do
      response = request('http://cl.ly/DPcp/Screen%20Shot%202012-01-17%20at%203.42.32%20PM.png')
      assert_equal(200, response.code)
    end

    should "follow redirects with path only location headers" do
      assert_nothing_raised do
        request('http://blogs.msdn.com/photos/noahric/images/9948044/425x286.aspx')
      end
    end

    should "404 on infinidirect" do
      assert_raise RestClient::ResourceNotFound do
        request('http://modeselektor.herokuapp.com/')
      end
    end

    should "404 on urls without an http host" do
      assert_raise RestClient::ResourceNotFound do
        request('/picture/Mincemeat/Pimp.jpg')
      end
    end

    should "404 on images larger than 5 MB" do
      assert_raise RestClient::ResourceNotFound do
        request('http://apod.nasa.gov/apod/image/0505/larryslookout_spirit_big.jpg')
      end
    end

    should "404 on host not found" do
      assert_raise RestClient::ResourceNotFound do
        request('http://flabergasted.cx')
      end
    end

    should "404 on non image content type" do
      assert_raise RestClient::ResourceNotFound do
        request('https://github.com/atmos/cinderella/raw/master/bootstrap.sh')
      end
    end

    should "404 on 10.0 ip range" do
      assert_raise RestClient::ResourceNotFound do
        request('http://10.0.0.1/foo.cgi')
      end
    end

    16.upto(31) do |i|
      should "test 404s on 172.#{i} ip range" do
        assert_raise RestClient::ResourceNotFound do
          request("http://172.#{i}.0.1/foo.cgi")
        end
      end
    end

    should "404 on 169.254 ip range" do
      assert_raise RestClient::ResourceNotFound do
        request('http://169.254.0.1/foo.cgi')
      end
    end

    should "404 on 192.168 ip range" do
      assert_raise RestClient::ResourceNotFound do
        request('http://192.168.0.1/foo.cgi')
      end
    end

    should "404 on excludes" do
      assert_raise RestClient::ResourceNotFound do
        request('http://iphone.internal.example.org/foo.cgi')
      end
    end
  end
end
