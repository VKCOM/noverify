var Config = /** @class */ (function () {
    function Config() {
        this.enable_notice = true;
        this.enable_warning = true;
        this.enable_error = true;
        this.strict_mixed = false;
        this.enable_php7 = false;
        this.unused_var = "^_$";
    }
    return Config;
}());
/**
 * WasmConfigurator implements configurations using Wasm.
 */
var WasmConfigurator = /** @class */ (function () {
    function WasmConfigurator() {
        this.wasm = GolangWasmSingleton.getInstance();
        this.changeConfig(new Config());
    }
    WasmConfigurator.prototype.changeConfig = function (config) {
        this.config = config;
        // We need to change the shared variables so that we can read the new config from Go.
        this.wasm.props.configJson = JSON.stringify(this.config);
        if (this.onChange !== undefined) {
            this.onChange(this.config);
        }
    };
    WasmConfigurator.prototype.getConfig = function () {
        return this.config;
    };
    return WasmConfigurator;
}());
var SeverityType;
(function (SeverityType) {
    SeverityType[SeverityType["Error"] = 0] = "Error";
    SeverityType[SeverityType["Warning"] = 1] = "Warning";
    SeverityType[SeverityType["Notice"] = 2] = "Notice";
})(SeverityType || (SeverityType = {}));
var Severity = /** @class */ (function () {
    function Severity() {
    }
    Severity.getString = function (type) {
        switch (type) {
            case SeverityType.Error:
                return "ERROR";
            case SeverityType.Warning:
                return "WARNING";
            case SeverityType.Notice:
                return "NOTICE";
        }
    };
    return Severity;
}());
/**
 * Report class is responsible for storing one report.
 */
var Report = /** @class */ (function () {
    function Report(message, check_name, severity, from, to) {
        this.message = "";
        this.check_name = "";
        this.severity = SeverityType.Notice;
        this.message = message;
        this.check_name = check_name;
        this.severity = severity;
        this.from = from;
        this.to = to;
    }
    return Report;
}());
/**
 * reportToHTML returns an HTML representation of the report.
 * @param report Report
 */
function reportToHTML(report) {
    var docsBaseLink = 'https://github.com/VKCOM/noverify/blob/master/docs/checkers_doc.md#';
    var checkerDoc = docsBaseLink + report.check_name;
    var severity = Severity.getString(report.severity);
    var checkerName = " <a target=\"_blank\" href=\"" + checkerDoc + "-checker\">" + report.check_name + "</a>: ";
    var message = report.message;
    var reportLine = "\n<a class=\"report-line js-line\" \n    data-line=\"" + (report.from.line + 1) + "\" \n    data-char=\"" + report.from.ch + "\" \n    title=\"Go to line " + (report.from.line + 1) + "\">\n    line " + (report.from.line + 1) + ":" + report.from.ch + "\n</a>";
    return severity + checkerName + message + " at " + reportLine;
}
/**
 * WasmReporter is class implements receiving reports using Wasm.
 */
var WasmReporter = /** @class */ (function () {
    function WasmReporter() {
        this.wasm = GolangWasmSingleton.getInstance();
    }
    WasmReporter.prototype.getReports = function (code) {
        if (this.wasm.props.analyzeCallback !== undefined) {
            this.wasm.props.analyzeCallback();
        }
        var reports;
        try {
            reports = JSON.parse(this.wasm.props.reportsJson);
        }
        catch (e) {
            console.error("Error parse reports json: ", this.wasm.props.reportsJson);
            return [];
        }
        if (reports === null) {
            return [];
        }
        return reports;
    };
    return WasmReporter;
}());
var Playground = /** @class */ (function () {
    function Playground(textArea, options, configurator, storage, reporter) {
        var _this = this;
        this.editor = CodeMirror.fromTextArea(textArea, options);
        this.configurator = configurator;
        this.storage = storage;
        this.reporter = reporter;
        // @ts-ignore
        options.lint.getAnnotations = function (code, callback) {
            var reports = _this.reporter.getReports(code);
            _this.onAnalyze(reports);
            callback(reports.map(function (report) {
                return {
                    message: report.message,
                    check_name: report.check_name,
                    severity: report.severity === SeverityType.Error ? 'error' : 'warning',
                    from: report.from,
                    to: report.to,
                };
            }));
        };
    }
    Playground.prototype.getCode = function () {
        return this.editor.getValue();
    };
    Playground.prototype.saveOnChange = function (val) {
        var _this = this;
        if (val) {
            this.editor.on("change", function () { return _this.storage.saveCode(_this.editor); });
        }
        else {
            this.editor.off("change", function () { return _this.storage.saveCode(_this.editor); });
        }
    };
    Playground.prototype.run = function () {
        var code = this.storage.getCode() || defaultCode;
        this.editor.setValue(code);
    };
    return Playground;
}());
var PlaygroundStorage = /** @class */ (function () {
    function PlaygroundStorage() {
    }
    PlaygroundStorage.prototype.saveCode = function (editor) {
        if (!editor) {
            return;
        }
        localStorage.setItem('noverify-playground-code', editor.getValue());
    };
    PlaygroundStorage.prototype.getCode = function () {
        return localStorage.getItem('noverify-playground-code');
    };
    return PlaygroundStorage;
}());
var GolangWasmSingleton = /** @class */ (function () {
    function GolangWasmSingleton() {
    }
    GolangWasmSingleton.getInstance = function () {
        if (!GolangWasmSingleton.instance) {
            GolangWasmSingleton.instance = new GolangWasm();
        }
        return GolangWasmSingleton.instance;
    };
    return GolangWasmSingleton;
}());
var GolangWasmProperties = /** @class */ (function () {
    function GolangWasmProperties() {
        this.reportsJson = '[]';
        this.configJson = '{}';
    }
    return GolangWasmProperties;
}());
/**
 * The GolangWasm class is responsible for starting a stream for wasm,
 * as well as storing fields that will be passed from JS to Golang and
 * vice versa.
 */
var GolangWasm = /** @class */ (function () {
    function GolangWasm() {
        // @ts-ignore
        this.go = new Go();
        this.props = new GolangWasmProperties();
    }
    GolangWasm.prototype.run = function (callback) {
        var _this = this;
        WebAssembly.instantiateStreaming(fetch('main.wasm'), this.go.importObject).then(function (result) {
            _this.go.run(result.instance);
            callback();
        });
    };
    return GolangWasm;
}());
var defaultCode = "<?php\n\nnamespace NoVerify;\n\nclass PlaygroundBase {\n    public abstract function analyze();\n}\n\n/**\n * @property $a\n */\nclass Playground extends PlaygroundBase {\n    use PlaygroundTrait;\n    \n    /** @var Analyzer */\n    var $analyzer = null;\n    /** @var callable(string): void */\n    var $cb = null;\n    \n    \n    /**\n     * @param Analyzer               $a\n     * @param callable(string): void $cb\n     * @param int                    $id\n     */\n    function __construct(Analyzer $a, callable $cb) {\n        $this->cb = $cb;\n        $analyzer = $analyzer;\n    }\n    \n    /** \n     * @see Plauground\n     * @return Reports[]\n     */\n    public function getReports(): array {\n        $callback = $this->cb;\n        \n        $warnings_count = 0;\n        $errors_count = 0;\n        $reports = array(\"\");\n        foreach ($reports as $index => $report) {\n            $hasReports = true;\n            \n            switch ($report[0]) {\n                case 'w':\n                    $warnings_count++;\n                    break;\n                case 'e':\n                    $warnings_count++;\n                    break;\n            }\n            $callback($report);\n        }\n       \n        $last_report = $reports[count($reports)];\n        \n        if (DEBUG) {\n            printf(\"Log: %s, time: %d, has %d\", (string)$last_report, $hasReports ?? false);\n        }\n        \n        return [$reports, $errors_count, $warnings_count];\n    }\n    \n    private function __set($name) {}\n    private function __get($name) {}\n}\n\n/**\n * @param array{obj:?Analyzer,id:int} $analyzers\n * @param callable(string): void      $cb\n */\n#[Pure]\nfunction runAnalyzers($analyzers, $cb) {\n    $analyzers[\"obj\"]->analyze();\n    $cb();\n}\n\nfunction main() {\n    $analyzers = [\"obj\" => new Analyzer(), \"id\" => 1];\n    $cb = function(string $v): void {};\n    \n    runAnalyzers($cb, $analyzers);\n}\n";
function bindOptions(configurator) {
    var config = configurator.getConfig();
    var optionsList = document.getElementsByClassName('options-list')[0];
    optionsList.addEventListener('change', function (e) {
        var target = e.target;
        var input = target;
        if (target.id === 'enable-notice') {
            config.enable_notice = input.checked;
        }
        else if (target.id === 'enable-warning') {
            config.enable_warning = input.checked;
        }
        else if (target.id === 'enable-error') {
            config.enable_error = input.checked;
        }
        else if (target.id === 'strict-mixed') {
            config.strict_mixed = input.checked;
        }
        else if (target.id === 'enable-php7') {
            config.enable_php7 = input.checked;
        }
        else if (target.id === 'unused-var') {
            config.unused_var = input.value;
        }
        else {
            return;
        }
        configurator.changeConfig(config);
    });
}
function bindButtons(editor) {
    var settingsButton = document.getElementsByClassName('settings')[0];
    settingsButton.addEventListener('click', function () {
        var reportsBlock = document.getElementsByClassName('reports')[0];
        reportsBlock.classList.toggle('open-options');
    });
    var minimizeButton = document.getElementsByClassName('js-minimize-button')[0];
    minimizeButton.addEventListener('click', function () {
        var reportsBlock = document.getElementsByClassName('reports')[0];
        reportsBlock.classList.remove('open-options');
        reportsBlock.classList.toggle('close');
    });
    var shareButton = document.getElementsByClassName('js-share-button')[0];
    shareButton.addEventListener('click', function () {
    });
    var reportsText = document.getElementsByClassName('reports-output-text')[0];
    reportsText.addEventListener('click', function (e) {
        var target = e.target;
        if (target.dataset.line !== undefined) {
            var line = parseInt(target.getAttribute('data-line'));
            var char = parseInt(target.getAttribute('data-char'));
            editor.focus();
            editor.setCursor({ line: line - 1, ch: char });
        }
    });
}
function main() {
    var editorTextArea = document.getElementById('editor');
    var editorConfig = {
        mode: 'php',
        lineNumbers: true,
        // @ts-ignore
        matchBrackets: true,
        extraKeys: { 'Ctrl-Space': 'autocomplete' },
        indentWithTabs: false,
        indentUnit: 4,
        autoCloseBrackets: true,
        showHint: true,
        lint: {
            async: true,
            lintOnChange: true,
            delay: 20,
        },
    };
    playground = new Playground(editorTextArea, editorConfig, new WasmConfigurator(), new PlaygroundStorage(), new WasmReporter());
    var editor = playground.editor;
    var configurator = playground.configurator;
    playground.saveOnChange(true);
    bindButtons(editor);
    bindOptions(configurator);
    var errorsForm = document.getElementsByClassName('reports-output-text')[0];
    if (errorsForm === undefined) {
        return;
    }
    playground.onAnalyze = function (reports) {
        console.log(reports);
        if (reports.length === 0) {
            errorsForm.innerHTML = 'No reports. Code is perfect!';
            return;
        }
        var criticalReports = reports.filter(function (x) {
            return x.severity === SeverityType.Error ||
                x.severity === SeverityType.Warning;
        }).length;
        var minorReports = reports.filter(function (x) {
            return x.severity === SeverityType.Notice;
        }).length;
        var HTMLReportsList = reports.map(function (report) { return reportToHTML(report); });
        var HTMLReports = HTMLReportsList.join("<br><br>");
        var header = "Found " + criticalReports + " critical and " + minorReports + " minor reports<br><br>";
        errorsForm.innerHTML = header + HTMLReports;
    };
    configurator.onChange = function () {
        // @ts-ignore
        editor.performLint();
    };
    GolangWasmSingleton.getInstance().run(function () {
        // @ts-ignore
        editor.performLint();
    });
    playground.run();
}
// For wasm.
var playground;
var wasm = GolangWasmSingleton.getInstance();
main();
