class DevStack < Formula
  desc "Development stack management tool for streamlined local development automation"
  homepage "https://github.com/isaacgarza/dev-stack"
  license "MIT"

  on_macos do
    on_intel do
      url "https://github.com/isaacgarza/dev-stack/releases/download/v1.0.0/dev-stack-darwin-amd64"
      sha256 "PLACEHOLDER_SHA256_AMD64"
    end

    on_arm do
      url "https://github.com/isaacgarza/dev-stack/releases/download/v1.0.0/dev-stack-darwin-arm64"
      sha256 "PLACEHOLDER_SHA256_ARM64"
    end
  end

  def install
    bin.install "dev-stack-darwin-#{Hardware::CPU.arch}" => "dev-stack"
  end

  test do
    assert_match "dev-stack", shell_output("#{bin}/dev-stack --version")
  end
end
