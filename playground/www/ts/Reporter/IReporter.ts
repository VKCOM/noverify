/**
 * IReporter interface is responsible for receiving
 * reports for the given code.
 */
interface IReporter {
  getReports(code: string): Report[];
}
