class TelosMatrix < Formula
  desc "Idea capture + Telos-aligned analysis for decision paralysis"
  homepage "https://github.com/your-repo/telos-idea-matrix"
  url "file:///Users/rayyacub/Documents/CCResearch/telos-idea-matrix"
  version "0.1.0"
  license "MIT"

  depends_on "rust" => :build
  depends_on "sqlite"
  depends_on "ollama" => :optional

  def install
    system "cargo", "build", "--release"
    bin.install "target/release/tm"
  end

  test do
    system "#{bin}/tm", "--version"
    system "#{bin}/tm", "--help"
  end
end