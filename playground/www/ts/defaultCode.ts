const defaultCode = `<?php

namespace NoVerify;

class PlaygroundBase {
    public abstract function analyze();
}

/**
 * @property $a
 */
class Playground extends PlaygroundBase {
    use PlaygroundTrait;
    
    /** @var Analyzer */
    var $analyzer = null;
    /** @var callable(string): void */
    var $cb = null;
    
    
    /**
     * @param Analyzer               $a
     * @param callable(string): void $cb
     * @param int                    $id
     */
    function __construct(Analyzer $a, callable $cb) {
        $this->cb = $cb;
        $analyzer = $analyzer;
    }
    
    /** 
     * @see Plauground
     * @return Reports[]
     */
    public function getReports(): array {
        $callback = $this->cb;
        
        $warnings_count = 0;
        $errors_count = 0;
        $reports = array("");
        foreach ($reports as $index => $report) {
            $hasReports = true;
            
            switch ($report[0]) {
                case 'w':
                    $warnings_count++;
                    break;
                case 'e':
                    $warnings_count++;
                    break;
            }
            $callback($report);
        }
       
        $last_report = $reports[count($reports)];
        
        if (DEBUG) {
            printf("Log: %s, time: %d, has %d", (string)$last_report, $hasReports ?? false);
        }
        
        return [$reports, $errors_count, $warnings_count];
    }
    
    private function __set($name) {}
    private function __get($name) {}
}

/**
 * @param array{obj:?Analyzer,id:int} $analyzers
 * @param callable(string): void      $cb
 */
function runAnalyzers($analyzers, $cb) {
    $analyzers["obj"]->analyze();
    $cb();
}

function main() {
    $analyzers = ["obj" => new Analyzer(), "id" => 1];
    $cb = function(string $v): void {};
    
    runAnalyzers($cb, $analyzers);
}
`
