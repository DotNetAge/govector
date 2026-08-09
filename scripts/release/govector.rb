class Govector < Formula
  desc "Lightweight, embeddable vector database in pure Go (Qdrant compatible)"
  homepage "https://github.com/__GITHUB_REPO__"
  version "__VERSION__"

  # MacOS ARM64 (M1/M2/M3)
  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/__GITHUB_REPO__/releases/download/v__VERSION__/govector_v__VERSION___darwin_arm64.tar.gz"
    sha256 "__SHA256_ARM64__"
  end

  # MacOS AMD64 (Intel)
  if OS.mac? && Hardware::CPU.intel?
    url "https://github.com/__GITHUB_REPO__/releases/download/v__VERSION__/govector_v__VERSION___darwin_amd64.tar.gz"
    sha256 "__SHA256_AMD64__"
  end

  # Linux
  if OS.linux? && Hardware::CPU.intel?
    url "https://github.com/__GITHUB_REPO__/releases/download/v__VERSION__/govector_v__VERSION___linux_amd64.tar.gz"
    sha256 "__SHA256_LINUX__"
  end

  def install
    bin.install "govector"

    # Create the data directory
    (var/"govector").mkpath
  end

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
