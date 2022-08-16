<?php

namespace NoVerify;

use Exception;
use ZipArchive;

class Downloader {
  /**
   * Path to releases on github.
   */
  private const BASE_PATH = 'https://github.com/VKCOM/noverify/releases/download/';

  /**
   * List of available versions for download.
   */
  public const VERSIONS = [
    "0.3.0",
    "0.4.0",
    "0.5.0",
    "0.5.1",
    "0.5.2",
    "0.5.3",
  ];

  /**
   * Returns the name of the current OS for choosing the correct binary.
   * @return string
   * @throws Exception If not supported OS.
   */
  private static function osName(): string {
    $name = php_uname('s');
    $name = strtolower($name);

    if (strpos($name, "windows") !== false) {
      return "windows";
    } elseif (strpos($name, "darwin") !== false) {
      return "darwin";
    } elseif (strpos($name, "linux") !== false) {
      return "linux";
    }

    throw new Exception("Not supported os: " . $name);
  }

  /**
   * Returns the architecture of the current processor for choosing the correct binary.
   * @return string
   * @throws Exception If not supported Arch.
   */
  private static function osArch(): string {
    $name = php_uname('m');

    if (strpos($name, "x86_64") !== false || strpos($name, "AMD64") !== false) {
      return "amd64";
    } elseif (strpos($name, "arm64") !== false) {
      return "arm64";
    }

    throw new Exception("Not supported arch: " . $name);
  }

  /**
   * Begins the process of downloading and unpacking the binary to the required folder.
   * @param string $version Version to process.
   */
  public static function process(string $version) {
    if ($version === "latest") {
      $version = self::VERSIONS[count(self::VERSIONS) - 1];
    }

    try {
      echo "Start download v$version version...\n";
      self::download($version);
      echo "Successful download v$version version\n";
      echo "Start extract v$version version...\n";
      self::extract($version);
      echo "Successful extracted v$version version\n";
    } catch (Exception $e) {
      echo $e->getMessage() . "\n";
    }
  }

  /**
   * Downloads the archive with the required version.
   * @param string $version Version to download
   * @throws Exception
   */
  public static function download(string $version) {
    if ($version === "0.2.0" || $version === "0.1.0") {
      throw new Exception("Version v$version cannot be downloaded");
    }

    $os   = self::osName();
    $arch = self::osArch();

    echo "Search version for OS: $os and Arch: $arch\n";

    $abs_path = self::BASE_PATH . "/v$version/noverify-$os-$arch.zip";

    $contents = @file_get_contents($abs_path);

    if ($contents === false && $arch === "arm64") {
      echo "Arm64 version not found\n";
      echo "Try download amd64 version\n";
      // Try download amd64 version.
      $abs_path = self::BASE_PATH . "/v$version/noverify-$os-amd64.zip";
      $contents = @file_get_contents($abs_path);
      if ($contents === false) {
        throw new Exception("Version v$version not found, available versions: " .
          join(", ", self::VERSIONS));
      }
      echo "Successful found amd64 version\n";
    } elseif ($contents === false) {
      throw new Exception("Version v$version not found, available versions: " .
        join(", ", self::VERSIONS));
    }

    @mkdir("vendor/bin");
    file_put_contents("./vendor/bin/noverify-$version.zip", $contents);
  }

  /**
   * Unpacks the previously downloaded archive into the ./vendor/bin folder.
   * And it makes the unpacked file executable.
   * @param string $version Version to extract
   * @throws Exception
   */
  public static function extract(string $version) {
    $os           = self::osName();
    $archive_name = "./vendor/bin/noverify-$version.zip";

    $zip = new ZipArchive;
    $res = $zip->open($archive_name);
    if ($res === false) {
      throw new Exception("Archive $archive_name not opened");
    }

    $zip->extractTo("./vendor/bin");
    if ($os != "windows") {
      system("chmod +x ./vendor/bin/noverify");
    }
  }
}
