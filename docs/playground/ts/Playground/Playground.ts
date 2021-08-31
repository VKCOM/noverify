class Playground {
  public readonly editor: CodeMirror.Editor
  public readonly configurator: IConfigurator
  public readonly storage: PlaygroundStorage
  public readonly reporter: IReporter

  public onAnalyze: (reports: Report[]) => void

  constructor(
    textArea: HTMLTextAreaElement,
    options: CodeMirror.EditorConfiguration,
    configurator: IConfigurator,
    storage: PlaygroundStorage,
    reporter: IReporter
  ) {
    this.editor = CodeMirror.fromTextArea(textArea, options)

    this.configurator = configurator
    this.storage = storage
    this.reporter = reporter

    // @ts-ignore
    options.lint.getAnnotations = (code: string, callback: (reports) => void) => {
      const reports = this.reporter.getReports(code)
      this.onAnalyze(reports)
      callback(reports.map((report: Report) => {
        return {
          message: report.message,
          check_name: report.check_name,
          severity: report.severity === SeverityType.Error ? 'error' : 'warning',
          from: report.from,
          to: report.to,
        }
      }))
    }
  }

  public getCode(): string {
    return this.editor.getValue()
  }

  public saveOnChange(val: boolean) {
    if (val) {
      this.editor.on("change", () => this.storage.saveCode(this.editor))
    } else {
      this.editor.off("change", () => this.storage.saveCode(this.editor))
    }
  }

  public run() {
    const code = this.storage.getCode() || defaultCode
    this.editor.setValue(code)
  }
}
