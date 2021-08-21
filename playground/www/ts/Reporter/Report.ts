enum SeverityType {
  Error,
  Warning,
  Notice
}

class Severity {
  public static getString(type: SeverityType): string {
    switch (type) {
      case SeverityType.Error:
        return "ERROR"
      case SeverityType.Warning:
        return "WARNING"
      case SeverityType.Notice:
        return "NOTICE"
    }
  }
}

/**
 * Report class is responsible for storing one report.
 */
class Report {
  public message: string = ""
  public check_name: string = ""
  public severity: SeverityType = SeverityType.Notice
  public from: CodeMirror.Position
  public to: CodeMirror.Position

  constructor(message: string, check_name: string, severity: SeverityType, from: CodeMirror.Position, to: CodeMirror.Position) {
    this.message = message;
    this.check_name = check_name;
    this.severity = severity;
    this.from = from;
    this.to = to;
  }
}

/**
 * reportToHTML returns an HTML representation of the report.
 * @param report Report
 */
function reportToHTML(report: Report): string {
  const docsBaseLink = 'https://github.com/VKCOM/noverify/blob/master/docs/checkers_doc.md#';
  const checkerDoc = docsBaseLink + report.check_name

  const severity = Severity.getString(report.severity)
  const checkerName = ` <a target="_blank" href="${checkerDoc}-checker">${report.check_name}</a>: `
  const message = report.message
  const reportLine = `
<a class="report-line js-line" 
    data-line="${report.from.line + 1}" 
    data-char="${report.from.ch}" 
    title="Go to line ${report.from.line + 1}">
    line ${report.from.line + 1}:${report.from.ch}
</a>`

  return severity + checkerName + message + " at " + reportLine
}
