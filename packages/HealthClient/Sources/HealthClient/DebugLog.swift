import Foundation

public struct DebugEvent: Codable, Sendable, Equatable {
  public var source: String
  public var level: String
  public var message: String
}

public struct DebugLog: Sendable {
  private let transport: any HealthTransport
  public init(transport: any HealthTransport = SessionTransport()) { self.transport = transport }

  public func report(address: String, level: String, message: String) async {
    do {
      let endpoint = try nestplayEndpoint(address: address, path: "/debug/log")
      var request = URLRequest(url: endpoint)
      request.httpMethod = "POST"
      request.cachePolicy = .reloadIgnoringLocalCacheData
      request.timeoutInterval = 5
      request.setValue("application/json", forHTTPHeaderField: "Content-Type")
      request.httpBody = try JSONEncoder().encode(
        DebugEvent(source: "tvos", level: level, message: message))
      _ = try await transport.load(request)
    } catch is CancellationError {
      return
    } catch {
      return
    }
  }
}
