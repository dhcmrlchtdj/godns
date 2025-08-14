class Godns < Formula
  desc "DNS server"
  homepage "https://github.com/dhcmrlchtdj/godns"
  license "Apache-2.0"
  head "https://github.com/dhcmrlchtdj/godns.git", branch: "main"

  depends_on "go" => :build

  def install
    ENV["GOEXPERIMENT"] = "greenteagc,jsonv2"
    system "go", "build", *std_go_args, "./cmd/godns"
    (etc/"godns").mkpath
    etc.install "aur/config.json" => "godns/config.json"
  end

  service do
    run [opt_bin/"godns", "--conf", etc/"godns/config.json"]
    keep_alive true
    log_path var/"log/godns.log"
    error_log_path var/"log/godns.log"
  end

  test do
    system "true"
  end
end
