<?php
class Simple
{
    const PORT = 31337;
    const SERVER_KEY = 'SERVER';

    /** @var resource[] (client_id => stream) */
    private $streams = [];
    /** @var string[] (client_id => read buffer) */
    private $read_buf = [];

    /** @var string[] (client_id => write buffer) */
    private $write_buf = [];
    /** @var resource[] (client_id => stream from which to read) */
    private $read = [];
    /** @var resource[] (client_id => stream where to write) */
    private $write = [];

    /** @var int Total connection count */
    private $conn_count = 0;

    public function run()
    {
        $this->listen();
        echo "Entering main loop\n";
        $this->mainLoop();
    }

    protected function listen()
    {
        $port = self::PORT;
        $ip_port = "0.0.0.0:$port";
        $address = "tcp://$ip_port";

        $server = stream_socket_server($address, $errno, $errstr, STREAM_SERVER_BIND | STREAM_SERVER_LISTEN);
        if (!$server) {
            fwrite(STDERR, "stream_socket_server failed: $errno $errstr\n");
            exit(1);
        }

        $this->read[self::SERVER_KEY] = $server;
        echo "Listening on $address\n";
    }

    public function response($stream_id, $response)
    {
        $json_resp = json_encode($response);
        echo "stream$stream_id " . $json_resp . "\n";
        $this->write($stream_id, $json_resp . "\n");
    }

    public function write($stream_id, $buf)
    {
        $this->write_buf[$stream_id] .= $buf;

        if (!isset($this->write[$stream_id])) {
            $this->write[$stream_id] = $this->streams[$stream_id];
        }
    }

    public function accept($server)
    {
        echo "Accepting new connection\n";

        $client = stream_socket_accept($server, 1, $peername);
        $stream_id = ($this->conn_count++);
        if (!$client) {
            fwrite(STDERR, "Accept failed\n");
            return;
        }

        stream_set_read_buffer($client, 0);
        stream_set_write_buffer($client, 0);
        stream_set_blocking($client, 0);
        stream_set_timeout($client, 1);

        $this->read_buf[$stream_id] = '';
        $this->write_buf[$stream_id] = '';
        $this->read[$stream_id] = $this->streams[$stream_id] = $client;

        echo "Connected stream$stream_id: $peername\n";
    }

    private function disconnect($stream_id)
    {
        echo "Disconnect stream$stream_id\n";
        unset($this->read_buf[$stream_id], $this->write_buf[$stream_id]);
        unset($this->streams[$stream_id]);
        unset($this->write[$stream_id], $this->read[$stream_id]);
    }

    private function handleRead($stream_id)
    {
        $buf = fread($this->streams[$stream_id], 8192);
        if ($buf === false || $buf === '') {
            echo "got EOF from stream$stream_id\n";
            if (empty($this->write_buf[$stream_id])) {
                $this->disconnect($stream_id);
            } else {
                unset($this->read[$stream_id]);
            }
            return;
        }

        $this->read_buf[$stream_id] .= $buf;
        $this->processJSONRequests($stream_id);
    }

    private function processJSONRequests($stream_id)
    {
        if (!strpos($this->read_buf[$stream_id], "\n")) return;
        $requests = explode("\n", $this->read_buf[$stream_id]);
        $this->read_buf[$stream_id] = array_pop($requests);

        foreach ($requests as $req) {
            $res = json_decode(rtrim($req), true);

            if ($res !== false) {
                $this->response($stream_id, "Request had " . count($res) . " keys");
            } else {
                $this->response($stream_id, "Invalid JSON");
            }
        }
    }

    private function handleWrite($stream_id)
    {
        if (!isset($this->write_buf[$stream_id])) {
            return;
        }

        $wrote = fwrite($this->streams[$stream_id], substr($this->write_buf[$stream_id], 0, 65536));
        if ($wrote === false) {
            fwrite(STDERR, "write failed into stream #$stream_id\n");
            $this->disconnect($stream_id);
            return;
        }

        if ($wrote === strlen($this->write_buf[$stream_id])) {
            $this->write_buf[$stream_id] = '';
            unset($this->write[$stream_id]);
            if (empty($this->read[$stream_id])) {
                $this->disconnect($stream_id);
            }
        } else {
            $this->write_buf[$stream_id] = substr($this->write_buf[$stream_id], $wrote);
        }
    }

    public function mainLoop()
    {
        while (true) {
            $read = $this->read;
            $write = $this->write;
            $except = null;

            echo "Selecting for " . count($read) . " reads, " . count($write) . " writes\n";
            $n = stream_select($read, $write, $except, NULL);

            if (!$n) {
                fwrite(STDERR, "Could not stream_select()\n");
            }

            if (count($read)) {
                echo "Can read from " . count($read) . " streams\n";
            }

            if (count($write)) {
                echo "Can write to " . count($write) . " streams\n";
            }

            if (isset($read[self::SERVER_KEY])) {
                $this->accept($read[self::SERVER_KEY]);
                unset($read[self::SERVER_KEY]);
            }

            foreach ($read as $stream_id => $_) {
                $this->handleRead($stream_id);
            }

            foreach ($write as $stream_id => $_) {
                $this->handleWrite($stream_id);
            }
        }
    }
}

$instance = new Simple();
$instance->run();
