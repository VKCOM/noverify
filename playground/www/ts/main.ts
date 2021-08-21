function bindOptions(configurator: IConfigurator) {
  const config = configurator.getConfig()

  const optionsList = document.getElementsByClassName('options-list')[0];
  optionsList.addEventListener('change', (e) => {
    const target = e.target as HTMLElement
    const input = target as HTMLInputElement

    if (target.id === 'enable-notice') {
      config.enable_notice = input.checked;
    } else if (target.id === 'enable-warning') {
      config.enable_warning = input.checked;
    } else if (target.id === 'enable-error') {
      config.enable_error = input.checked;
    } else if (target.id === 'strict-mixed') {
      config.strict_mixed = input.checked;
    } else if (target.id === 'enable-php7') {
      config.enable_php7 = input.checked;
    } else if (target.id === 'unused-var') {
      config.unused_var = input.value;
    } else {
      return;
    }

    configurator.changeConfig(config);
  });
}

function bindButtons(editor: CodeMirror.Editor) {
  const settingsButton = document.getElementsByClassName('settings')[0];
  settingsButton.addEventListener('click', function () {
    const reportsBlock = document.getElementsByClassName('reports')[0];
    reportsBlock.classList.toggle('open-options');
  });

  const minimizeButton = document.getElementsByClassName('js-minimize-button')[0];
  minimizeButton.addEventListener('click', function () {
    const reportsBlock = document.getElementsByClassName('reports')[0];
    reportsBlock.classList.remove('open-options');
    reportsBlock.classList.toggle('close');
  });

  const shareButton = document.getElementsByClassName('js-share-button')[0];
  shareButton.addEventListener('click', () => {
  });

  const reportsText = document.getElementsByClassName('reports-output-text')[0];
  reportsText.addEventListener('click', function (e) {
    const target = e.target as HTMLElement;
    if (target.dataset.line !== undefined) {
      const line = parseInt(target.getAttribute('data-line'));
      const char = parseInt(target.getAttribute('data-char'));
      editor.focus();
      editor.setCursor({line: line - 1, ch: char});
    }
  });
}

function main() {
  const editorTextArea = document.getElementById('editor') as HTMLTextAreaElement;

  const editorConfig: CodeMirror.EditorConfiguration = {
    mode: 'php',
    lineNumbers: true,
    // @ts-ignore
    matchBrackets: true,
    extraKeys: {'Ctrl-Space': 'autocomplete'},
    indentWithTabs: false,
    indentUnit: 4,
    autoCloseBrackets: true,
    showHint: true,
    lint: {
      async: true,
      lintOnChange: true,
      delay: 20,
    },
  }

  playground = new Playground(
    editorTextArea, editorConfig,
    new WasmConfigurator(),
    new PlaygroundStorage(),
    new WasmReporter(),
  )

  const editor = playground.editor
  const configurator = playground.configurator

  playground.saveOnChange(true)

  bindButtons(editor)
  bindOptions(configurator)

  const errorsForm = document.getElementsByClassName('reports-output-text')[0];
  if (errorsForm === undefined) {
    return;
  }

  playground.onAnalyze = (reports: Report[]) => {
    console.log(reports)

    if (reports.length === 0) {
      errorsForm.innerHTML = 'No reports. Code is perfect!'
      return
    }

    const criticalReports = reports.filter(
      x =>
        x.severity === SeverityType.Error ||
        x.severity === SeverityType.Warning
    ).length

    const minorReports = reports.filter(
      x =>
        x.severity === SeverityType.Notice
    ).length

    const HTMLReportsList = reports.map((report: Report) => reportToHTML(report))
    const HTMLReports = HTMLReportsList.join("<br><br>")

    const header = `Found ${criticalReports} critical and ${minorReports} minor reports<br><br>`
    errorsForm.innerHTML = header + HTMLReports
  }

  configurator.onChange = () => {
    // @ts-ignore
    editor.performLint()
  }
  GolangWasmSingleton.getInstance().run(() => {
    // @ts-ignore
    editor.performLint();
  })

  playground.run()
}

// For wasm.
var playground: Playground
var wasm = GolangWasmSingleton.getInstance()

main()
