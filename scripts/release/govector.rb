class Govector < Formula
  desc "Lightweight, embeddable vector database in pure Go (Qdrant compatible)"
  homepage "https://github.com/DotNetAge/govector"
  version "0.1.4"

  # MacOS ARM64 (M1/M2/M3)
  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/DotNetAge/govector/releases/download/v0.1.4/govector_v0.1.4_darwin_arm64.tar.gz"
    # To get the SHA256: run `shasum -a 256 govector_v0.1.4_darwin_arm64.tar.gz`
    sha256 "bbdfb934b9cd7eaa21c9a64131ea08fb14db06d719fe1e350145ca189c1766d3"
  end
  
  # MacOS AMD64 (Intel)
  if OS.mac? && Hardware::CPU.intel?
    url "https://github.com/DotNetAge/govector/releases/download/v0.1.4/govector_v0.1.4_darwin_amd64.tar.gz"
    sha256 "759d28c9ad3a2c0ccf05ff0d97b0f6d0978da7884c1a9330ee871a3b18e34cfa"
  end

  # Linux
  if OS.linux? && Hardware::CPU.intel?
    url "https://github.com/DotNetAge/govector/releases/download/v0.1.4/govector_v0.1.4_linux_amd64.tar.gz"
    sha256 "6d3acf1d61dded99576b0ee0fcb0722f8c114a0b380fdf32aef03704d2db781b"
  end

  def install
    bin.install "govector"
    
    # Create the data directory
    (var/"govector").mkpath
  end

  # This makes 'brew services start govector' work beautifully!
  service do
    run [opt_bin/"govector", "serve", "-port", "18080", "-db", var/"govector/data.db"]
    keep_alive true
    log_path var/"log/govector.log"
    error_log_path var/"log/govector_error.log"
    working_dir var/"govector"
  end

  test do
    system "#{bin}/govector", "-h"
  end
end
