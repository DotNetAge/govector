class Govector < Formula
  desc "Lightweight, embeddable vector database in pure Go (Qdrant compatible)"
  homepage "https://github.com/yourusername/govector"
  version "0.1.0"

  # MacOS ARM64 (M1/M2/M3)
  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/yourusername/govector/releases/download/v0.1.0/govector_v0.1.0_darwin_arm64.tar.gz"
    # To get the SHA256: run `shasum -a 256 govector_v0.1.0_darwin_arm64.tar.gz`
    sha256 "REPLACE_WITH_ACTUAL_SHA256"
  end
  
  # MacOS AMD64 (Intel)
  if OS.mac? && Hardware::CPU.intel?
    url "https://github.com/yourusername/govector/releases/download/v0.1.0/govector_v0.1.0_darwin_amd64.tar.gz"
    sha256 "REPLACE_WITH_ACTUAL_SHA256"
  end

  # Linux
  if OS.linux? && Hardware::CPU.intel?
    url "https://github.com/yourusername/govector/releases/download/v0.1.0/govector_v0.1.0_linux_amd64.tar.gz"
    sha256 "REPLACE_WITH_ACTUAL_SHA256"
  end

  def install
    bin.install "govector" => "govectord"
    
    # Create the data directory
    (var/"govector").mkpath
  end

  # This makes 'brew services start govector' work beautifully!
  service do
    run [opt_bin/"govectord", "-port", "18080", "-db", var/"govector/data.db", "-hnsw=true"]
    keep_alive true
    log_path var/"log/govector.log"
    error_log_path var/"log/govector_error.log"
    working_dir var/"govector"
  end

  test do
    system "#{bin}/govectord", "-h"
  end
end
