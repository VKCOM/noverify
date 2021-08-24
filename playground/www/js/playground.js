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
var defaultCode = "<?php\n\nclass FooWithFinalMethod {\n    final function f() {}\n}\n\nclass BooWithSameMethod extends FooWithFinalMethod {\n    function f() {}\n}\n\nabstract class AbstractClass {\n    abstract public function abstractMethod() {}\n}\n\nclass SomeClass extends AbstractClass {}\n\n/**\n * @method int a\n */\nclass Boo {}\n\n/**\n * @method void check()\n */\nfinal class Foo {\n    var $prop, $prop2;\n  \n    /**\n     * @var Boo\n     */\n    var $p = null;\n\n    /**\n     * Instance method\n     */\n    function instanceMethod(int $x) {}\n    \n    final public static function staticMethod(int $x) { \n        echo $this->p;\n    }\n\n    public function __call($name) {}\n}\n\n/**\n * @param  array{int,Foo} $x1\n * @return array{int,Foo}\n */\nfunction getArray(array $x) { return [0, new Foo]; }\n\n/**\n * @param callable(int) $a\n * @param callable(int): Foo $b\n */\nfunction mainCheck(callable $a, callable $b) {\n    echo getArray();\n    (new Foo)->instanceMethod();\n\n    echo getArray(10)[1]->p->f;\n\n    echo (new Foo)->check();\n\n    /**\n     * @return callable(int, string): Foo\n     */\n    $b = function (int $a) { };\n    $c = $b(10);\n    $c();\n}\n\nfunction makeHello(string $name, int $age) {\n    echo \"Hello ${$name}-${$age1}\";\n}\n\nfunction main(): void {\n    $name = \"John\";\n    $age = 18;\n    echo makeHello($age, $name);\n}\n";
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
