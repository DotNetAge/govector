class Govector < Formula
  desc "Lightweight, embeddable vector database in pure Go (Qdrant compatible)"
  homepage "https://github.com/DotNetAge/govector"
  version "0.1.3"

  # MacOS ARM64 (M1/M2/M3)
  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/DotNetAge/govector/releases/download/v0.1.3/govector_v0.1.3_darwin_arm64.tar.gz"
    # To get the SHA256: run `shasum -a 256 govector_v0.1.3_darwin_arm64.tar.gz`
    sha256 "76e91e5c9581eeff87ef959a21bce5f4f8cd196bfd477c703732bc4ac5c3862d"
  end
  
  # MacOS AMD64 (Intel)
  if OS.mac? && Hardware::CPU.intel?
    url "https://github.com/DotNetAge/govector/releases/download/v0.1.3/govector_v0.1.3_darwin_amd64.tar.gz"
    sha256 "2f2ff257b0d86d0b0225d55438f695a4c807ff631ac73848ee4a7271fdcf7458"
  end

  # Linux
  if OS.linux? && Hardware::CPU.intel?
    url "https://github.com/DotNetAge/govector/releases/download/v0.1.3/govector_v0.1.3_linux_amd64.tar.gz"
    sha256 "7a1dca7ec3082b461c07b0b2eb5b42cef02d973ca8a0e1df1b25bcd0efdf0d73"
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
