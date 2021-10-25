package checkers

import (
	"testing"

	"github.com/VKCOM/noverify/src/linttest"
)

func TestPackagingArgs(t *testing.T) {
	test := linttest.NewSuite(t)
	test.AddNamedFile("src/HttpClient/UnixSocket.php", `<?php
namespace HttpClient;

/**
 * @package HttpClient
 */
class UnixSocket {
  /**
   * @internal
   */
  public static function doGet(string $url): string {}
}
`)
	test.AddNamedFile("src/HttpClient/Socket.php", `<?php
namespace HttpClient;

/**
 * @package HttpClient
 * @internal
 */
class Socket {
  public static function doGet(string $url): string {}
}
`)
	test.AddNamedFile("src/HttpClient/JsonSocket.php", `<?php
namespace HttpClient;

/**
 * @package HttpClient
 */
class JsonSocket {
  public static function doGet(string $url): string {
  	return Socket::doGet($url);
  }
}
`)
	test.AddNamedFile("cmd/main.php", `<?php
require_once '../src/HttpClient/Socket.php';
require_once '../src/HttpClient/UnixSocket.php';
require_once '../src/HttpClient/JsonSocket.php';

class Main {
  public static function fetch(): void {
	\HttpClient\Socket::doGet('https://example.com');

	\HttpClient\UnixSocket::doGet('https://example.com');
	
	\HttpClient\JsonSocket::doGet('https://example.com');
  }
}
`)
	test.Expect = []string{
		`Call @internal method \HttpClient\Socket::doGet outside package HttpClient`,
		`Call @internal method \HttpClient\UnixSocket::doGet outside package HttpClient`,
	}
	linttest.RunFilterMatch(test, "packaging")
}
