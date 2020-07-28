<?php
class TestClass1
{
    public function run($args)
    {
        echo "Running as " . __CLASS__ . "\n";
        var_dump($args);
        sleep(10);
    }
}

class TestClass2
{
    public function run($args)
    {
        echo "Running as " . __CLASS__ . "\n";
        var_dump($args);
        sleep(100);
    }
}

class ScriptResult
{
    public
        $retcode,
        $is_running,
        $pid;
}

class Phprocksyd
{
    const CONN_TIMEOUT = 20;
    const ACCEPT_TIMEOUT = 1;  // How long can accept() be running before timing out

    const RESTART_FILENAME = 'phprocksyd.restart';
    const RESTART_DIR = '/tmp/';

    const SOCKET_BACKLOG = 1024;
    const PORT = 31337;

    const SERVER_KEY = 'SERVER';

    private static $restart_props = ['hash_info', 'pid_to_hash'];
    private static $restart_fd_resources = ['read', 'write', 'streams'];
    private static $restart_fd_props = ['read_buf', 'write_buf', 'conn_count'];

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

    /** @var array (pid => hash) */
    private $pid_to_hash = [];
    /** @var ScriptResult[] (hash => ScriptResult) */
    private $hash_info;

    /** @var int Total connection count */
    private $conn_count = 0;

    /** @var array we need to be able to close all "extra" file descriptors before exec so that we do not leak anything */
    private $known_fds = [0 => true, 1 => true, 2 => true];

    public function init()
    {
        foreach (['posix', 'pcntl'] as $ext) {
            if (!extension_loaded($ext) && !dl($ext . '.so')) {
                fwrite(STDERR, "Could not load $ext.so\n");
                exit(1);
            }
        }

        $pid = pcntl_fork();
        if ($pid < 0) {
            fwrite(STDERR, "Could not fork\n");
            exit(1);
        }

        if ($pid == 0) {
            usleep(1000);
            exit(0);
        }

        $metadata = stream_get_meta_data(STDIN);
        if (!$metadata) {
            fwrite(STDERR, "stream_get_metadata broken: false returned for STDIN\n");
            exit(1);
        }

        if (!isset($metadata['fd'])) {
            fwrite(STDERR, "stream_get_metadata broken: fd not present for STDIN\n");
            exit(1);
        }
    }

    public function run()
    {
        $this->loadRestartFile();
        $this->listen();

        echo "Entering main loop\n";
        $this->mainLoop();
        return true;
    }

    protected function listen()
    {
        $ctx = stream_context_create();
        $res = stream_context_set_option($ctx, 'socket', 'backlog', self::SOCKET_BACKLOG);
        if (!$res) {
            fwrite(STDERR, "Could not set backlog option on context\n");
            exit(1);
        }

        $port = self::PORT;
        $ip_port = "0.0.0.0:$port";
        $address = "tcp://$ip_port";

        if (!isset($this->read[self::SERVER_KEY])) {
            $server = stream_socket_server($address, $errno, $errstr, STREAM_SERVER_BIND | STREAM_SERVER_LISTEN, $ctx);
            if (!$server) {
                fwrite(STDERR, "stream_socket_server failed: $errno $errstr\n");
                exit(1);
            }

            $this->read[self::SERVER_KEY] = $server;
            echo "Listening on $address\n";
        } else {
            echo "Reusing listen socket for $address\n";
        }
    }

    protected function command($stream_id, $msg_name, $res, $res_json)
    {
        echo "stream$stream_id " . $msg_name . " " . $res_json . "\n";

        switch ($msg_name) {
            case 'request_check':
                $this->requestCheck($stream_id, $res);
                break;

            case 'request_free':
                $this->requestFree($stream_id, $res);
                break;

            case 'request_terminate':
                $this->requestTerminate($stream_id, $res);
                break;

            case 'request_run':
                $this->requestRun($stream_id, $res);
                break;

            case 'request_restart':
                $this->requestRestart($stream_id, $res);
                break;

            case 'request_stop':
                $this->requestStop($stream_id, $res);
                break;

            case 'request_bye':
                $this->requestBye($stream_id, $res);
                break;

            default:
                $msg = "UNKNOWN REQUEST '$msg_name': $res_json";
                echo $msg . "\n";
                $this->generic($stream_id, $msg);
                return;
        }
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

        $client = stream_socket_accept($server, self::ACCEPT_TIMEOUT, $peername);
        $stream_id = ($this->conn_count++);
        if (!$client) {
            fwrite(STDERR, "Accept failed\n");
            return;
        }

        stream_set_read_buffer($client, 0);
        stream_set_write_buffer($client, 0);
        stream_set_blocking($client, 0);
        stream_set_timeout($client, self::CONN_TIMEOUT);

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

    private function cleanChildren()
    {
        while (true) {
            $status = null;
            $pid = pcntl_wait($status, WNOHANG);
            if ($pid <= 0) {
                return;
            }

            if (isset($this->pid_to_hash[$pid])) {
                $hash = $this->pid_to_hash[$pid];
                unset($this->pid_to_hash[$pid]);

                if ($Res = $this->hash_info[$hash]) {
                    $Res->retcode = pcntl_wexitstatus($status);
                    echo "pid $pid (hash $hash) wexit = " . $Res->retcode . "\n";
                }
            }
        }
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
            $req = rtrim($req);
            $parts = explode(" ", $req, 2);
            if (count($parts) != 2) {
                $parts[1] = '{}';
            }
            list($msg_name_start, $msg) = $parts;
            $msg_name = 'request_' . $msg_name_start;
            $res = json_decode($msg, true);

            if ($res !== false) {
                $this->command($stream_id, $msg_name, $res, $msg);
            } else {
                $this->generic($stream_id, 'Invalid JSON');
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
                exit(1);
            }

            if (count($read)) {
                echo "Can read from " . count($read) . " streams\n";
            }

            if (count($write)) {
                echo "Can write to " . count($write) . " streams\n";
            }

            $this->cleanChildren();

            if (isset($read[self::SERVER_KEY])) {
                $this->accept($read[self::SERVER_KEY]);
                unset($read[self::SERVER_KEY]);
            }

            // get rid of references to connection resources so that connections do not leak to children
            $read_keys = array_keys($read);
            $write_keys = array_keys($write);
            unset($read, $write);

            foreach ($read_keys as $stream_id) {
                $this->handleRead($stream_id);
            }

            foreach ($write_keys as $stream_id) {
                $this->handleWrite($stream_id);
            }
        }
    }

    private function msg_length($buf)
    {
        return ord($buf[0]) << 24 | ord($buf[1]) << 16 | ord($buf[2]) << 8 | ord($buf[3]);
    }

    /**
     * @param $stream_id
     * @param $req
     */
    private function requestFree($stream_id, $req)
    {
        $hash = $req['hash'];

        if (!isset($this->hash_info[$hash])) {
            $this->generic($stream_id, "Hash not found");
            return;
        }

        $this->free($hash);
        $this->generic($stream_id, "OK");
    }

    private function free($hash)
    {
        if (isset($this->hash_info[$hash])) {
            $Res = $this->hash_info[$hash];
            unset($this->pid_to_hash[$Res->pid], $this->hash_info[$hash]);
        }
    }

    /**
     * @param $stream_id
     * @param $req
     */
    private function requestCheck($stream_id, $req)
    {
        $status = null;
        $hash = $req['hash'];

        if (!is_numeric($hash)) {
            $this->generic($stream_id, "Hash is not numeric");
            return;
        }

        if (!isset($this->hash_info[$hash])) {
            echo "hash = $hash miss\n";
            $this->generic($stream_id, "Hash not found");
            return;
        }

        $Result = $this->hash_info[$hash];

        if (!isset($Result->retcode)) {
            echo "hash = $hash didn't exit\n";
            $this->generic($stream_id, "Still running");
            return;
        }

        $resp = ['retcode' => $Result->retcode];
        $this->response($stream_id, $resp);
    }

    /**
     * @param $stream_id
     * @param $req
     */
    protected function requestTerminate($stream_id, $req)
    {
        if (!isset($this->hash_info[$req['hash']])) {
            $this->generic($stream_id, "Hash not found");
            return;
        }

        $pid = $this->hash_info[$req['hash']]->pid;

        if (!isset($this->pid_to_hash[$pid])) {
            $this->generic($stream_id, "Script already not running");
            return;
        }

        $result = posix_kill($pid, SIGTERM);
        $this->generic($stream_id, $result ? "OK" : "Failed to kill");
    }

    /**
     * @param $stream_id
     * @param $req
     * @throws \Exception
     */
    protected function requestRun($stream_id, $req)
    {
        $hash = $req['hash'];
        if (!is_numeric($hash)) {
            $this->generic($stream_id, "hash is not numeric");
            return;
        }

        $pid = pcntl_fork();
        if ($pid == -1) {
            fwrite(STDERR, "Cannot fork\n");
            $this->generic($stream_id, "Cannot fork");
            return;
        }

        if ($pid == 0) {
            try {
                /* clean memory and close all parent connections */
                $this->streams = null;
                $this->read = null;
                $this->write = null;

                $args = array_merge([$req['class']], $req['params']);
                $title = "php " . implode(" ", $args);
                cli_set_process_title($title);

                $seed = floor(explode(" ", microtime())[0] * 1e6);
                srand($seed);
                mt_srand($seed);

                $class = $req['class'];
                $instance = new $class;
                $instance->run($req['params']);
            } finally {
                exit(0);
            }
        }

        echo "hash " . $hash . " ran as pid $pid\n";

        $res = new ScriptResult;
        $res->is_running = true;
        $res->pid = $pid;

        $this->hash_info[$hash] = $res;
        $this->pid_to_hash[$pid] = $hash;

        $this->generic($stream_id, "OK");
    }

    protected function generic($stream_id, $error_text)
    {
        $params = ['error_text' => $error_text];
        $this->response($stream_id, $params);
    }

    protected function requestRestart($stream_id)
    {
        echo "Restarting...\n";
        $this->generic($stream_id, "Restarted successfully");
        $this->restart();
    }

    protected function requestBye($stream_id)
    {
        $this->disconnect($stream_id);
    }

    private function getFdRestartData()
    {
        $res = [];

        foreach (self::$restart_fd_resources as $prop) {
            $res[$prop] = [];
            foreach ($this->$prop as $k => $v) {
                $meta = stream_get_meta_data($v);
                if (!isset($meta['fd'])) {
                    fwrite(STDERR, "No fd in stream metadata for resource $v (key $k in $prop), got " . var_export($meta, true) . "\n");
                    return false;
                }
                $res[$prop][$k] = $meta['fd'];
                $this->known_fds[$meta['fd']] = true;
            }
        }

        foreach (self::$restart_fd_props as $prop) {
            $res[$prop] = $this->$prop;
        }

        return $res;
    }

    private function loadFdRestartData($res)
    {
        $fd_resources = [];

        foreach (self::$restart_fd_resources as $prop) {
            if (!isset($res[$prop])) {
                fwrite(STDERR, "Property '$prop' is not present in restart fd resources\n");
                continue;
            }

            $pp = [];
            foreach ($res[$prop] as $k => $v) {
                if (isset($fd_resources[$v])) {
                    $pp[$k] = $fd_resources[$v];
                } else {
                    $fp = fopen("php://fd/" . $v, 'r+');
                    if (!$fp) {
                        fwrite(STDERR, "Failed to open fd = $v, exiting\n");
                        exit(1);
                    }

                    stream_set_read_buffer($fp, 0);
                    stream_set_write_buffer($fp, 0);
                    stream_set_blocking($fp, 0);
                    stream_set_timeout($fp, self::CONN_TIMEOUT);

                    $fd_resources[$v] = $fp;
                    $pp[$k] = $fp;
                }
            }
            $this->$prop = $pp;
        }

        foreach (self::$restart_fd_props as $prop) {
            if (!isset($res[$prop])) {
                fwrite(STDERR, "Property '$prop' is not present in restart fd properties\n");
                continue;
            }

            $this->$prop = $res[$prop];
        }
    }

    private function restart()
    {
        echo "Creating restart file...\n";
        $res = [];
        foreach (self::$restart_props as $prop) $res[$prop] = $this->$prop;

        if (!$add_res = $this->getFdRestartData()) {
            fwrite(STDERR, "Could not get restart FD data, exiting, graceful restart is not supported\n");
            exit(0);
        }

        $res += $add_res;

        /* Close all extra file descriptors that we do not know of, including opendir() descriptor :) */
        $dh = opendir("/proc/self/fd");
        $fds = [];
        while (false !== ($file = readdir($dh))) {
            if ($file[0] === '.') continue;
            $fds[] = $file;
        }

        foreach ($fds as $fd) {
            if (!isset($this->known_fds[$fd])) {
                fclose(fopen("php://fd/" . $fd, 'r+'));
            }
        }

        $contents = serialize($res);

        if (file_put_contents(self::RESTART_DIR . self::RESTART_FILENAME, $contents) !== strlen($contents)) {
            fwrite(STDERR, "Could not fully write restart file\n");
            unlink(self::RESTART_DIR . self::RESTART_FILENAME);
        }

        echo "Doing exec()\n";
        pcntl_exec("/usr/bin/env", ["php", __FILE__], $_ENV);
        fwrite(STDERR, "exec() failed\n");
        exit(1);
    }

    private function loadRestartFile()
    {
        if (!file_exists(self::RESTART_DIR . self::RESTART_FILENAME)) {
            return;
        }

        echo "Restart file found, trying to adopt it\n";

        $contents = file_get_contents(self::RESTART_DIR . self::RESTART_FILENAME);
        unlink(self::RESTART_DIR . self::RESTART_FILENAME);

        if ($contents === false) {
            fwrite(STDERR, "Could not read restart file\n");
            return;
        }

        $res = unserialize($contents);
        if (!$res) {
            fwrite(STDERR, "Could not unserialize restart file contents");
            return;
        }

        foreach (self::$restart_props as $prop) {
            if (!array_key_exists($prop, $res)) {
                fwrite(STDERR, "No property $prop in restart file\n");
                continue;
            }
            $this->$prop = $res[$prop];
        }

        $this->loadFdRestartData($res);

        // We could lose some children on the way, clean them up
        $this->cleanChildren();

        $our_pid = getmypid();
        $clean_pids = [];

        foreach ($this->pid_to_hash as $pid => $hash) {
            $filename = "/proc/$pid/stat";
            if (!file_exists($filename) || !is_readable($filename) || !($contents = file_get_contents($filename))) {
                $clean_pids[] = $pid;
                continue;
            }

            $parts = explode(' ', $contents);
            $ppid = $parts[3];
            if ($ppid != $our_pid) {
                $clean_pids[] = $pid;
            }
        }

        foreach ($clean_pids as $pid) {
            $hash = $this->pid_to_hash[$pid];
            unset($this->pid_to_hash[$pid]);

            echo "Lost info for pid = $pid, hash = " . $hash . "\n";
        }
    }

    private function requestStop($stream_id, $res)
    {
        echo "Requested to stop, exiting\n";
        exit(0);
    }
}

$instance = new Phprocksyd();
$instance->init();
$instance->run();
