class Govector < Formula
  desc "Lightweight, embeddable vector database in pure Go (Qdrant compatible)"
  homepage "https://github.com/DotNetAge/govector"
  version "0.1.0"

  # MacOS ARM64 (M1/M2/M3)
  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/DotNetAge/govector/releases/download/v0.1.0/govector_v0.1.0_darwin_arm64.tar.gz"
    # To get the SHA256: run `shasum -a 256 govector_v0.1.0_darwin_arm64.tar.gz`
    sha256 "3f9ff3f8bb33298dbeac6f6274258e7e4466afa4d054ccedb401a62e7ea869d2"
  end
  
  # MacOS AMD64 (Intel)
  if OS.mac? && Hardware::CPU.intel?
    url "https://github.com/DotNetAge/govector/releases/download/v0.1.0/govector_v0.1.0_darwin_amd64.tar.gz"
    sha256 "7a4d5e090f9f0f80efbd0f98282cec8f12d68ddc2f00c9e481054ec75ac431ab"
  end

  # Linux
  if OS.linux? && Hardware::CPU.intel?
    url "https://github.com/DotNetAge/govector/releases/download/v0.1.0/govector_v0.1.0_linux_amd64.tar.gz"
    sha256 "52de2412000a1cd46d2735d8827439fb7d8a91cfaff02c494f0af6223ce32455"
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
