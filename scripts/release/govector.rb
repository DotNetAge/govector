class Govector < Formula
  desc "Lightweight, embeddable vector database in pure Go (Qdrant compatible)"
  homepage "https://github.com/DotNetAge/govector"
  version "0.1.3"

  # MacOS ARM64 (M1/M2/M3)
  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/DotNetAge/govector/releases/download/v0.1.3/govector_v0.1.3_darwin_arm64.tar.gz"
    # To get the SHA256: run `shasum -a 256 govector_v0.1.3_darwin_arm64.tar.gz`
    sha256 "bad56fcd25813f73305b687afa3870bbb832b1f313fc21e250aabea6744ae8fe"
  end
  
  # MacOS AMD64 (Intel)
  if OS.mac? && Hardware::CPU.intel?
    url "https://github.com/DotNetAge/govector/releases/download/v0.1.3/govector_v0.1.3_darwin_amd64.tar.gz"
    sha256 "2f243d093b6b6163eeef9d61c725946e72a681e544ace00b623488646c601950"
  end

  # Linux
  if OS.linux? && Hardware::CPU.intel?
    url "https://github.com/DotNetAge/govector/releases/download/v0.1.3/govector_v0.1.3_linux_amd64.tar.gz"
    sha256 "7e874a12346d0a88ad6a0a713be14c3a7c0c6fd0447e15d1a4e38b58032444fc"
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
