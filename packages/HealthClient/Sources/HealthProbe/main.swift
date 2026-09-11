import Foundation
import HealthClient

do {
  let address = CommandLine.arguments.dropFirst().first ?? "http://127.0.0.1:8765"
  let result = try await HealthClient().check(address: address)
  print("Connected: \(result.service), API \(result.apiVersion), \(result.status)")
} catch {
  print(error.localizedDescription)
  exit(1)
}
