import Foundation
import Testing

@testable import HealthClient

private actor StubTransport: HealthTransport {
  let status: Int
  let body: String
  var request: URLRequest?
  init(
    status: Int = 200,
    body: String = "{\"service\":\"nestplay\",\"status\":\"ok\",\"api_version\":1}"
  ) {
    self.status = status
    self.body = body
  }
  func load(_ request: URLRequest) async throws -> (Data, URLResponse) {
    self.request = request
    return (
      Data(body.utf8),
      HTTPURLResponse(url: request.url!, statusCode: status, httpVersion: nil, headerFields: nil)!
    )
  }
}

@Test func checksContractAndEndpoint() async throws {
  let transport = StubTransport()
  let health = try await HealthClient(transport: transport).check(
    address: " http://mac.local:8765/ ")
  #expect(health.service == "nestplay")
  let request = await transport.request
  #expect(request?.url?.absoluteString == "http://mac.local:8765/health")
  #expect(request?.httpMethod == "GET")
}

@Test(arguments: [
  "mac.local", "file:///tmp/health", "http://mac.local/path", "http://name:secret@mac.local",
  "http://mac.local?x=1",
])
func rejectsInvalidAddresses(address: String) async {
  await #expect(throws: HealthError.invalidAddress) {
    try await HealthClient(transport: StubTransport()).check(address: address)
  }
}

@Test func rejectsWrongServiceAndVersion() async {
  for body in [
    "{\"service\":\"other\",\"status\":\"ok\",\"api_version\":1}",
    "{\"service\":\"nestplay\",\"status\":\"ok\",\"api_version\":2}",
  ] {
    await #expect(throws: HealthError.incompatibleServer) {
      try await HealthClient(transport: StubTransport(body: body)).check(
        address: "http://mac.local")
    }
  }
}

@Test func rejectsHTTPFailureAndMalformedResponse() async {
  for stub in [StubTransport(status: 503), StubTransport(body: "<html>not health</html>")] {
    await #expect(throws: HealthError.invalidResponse) {
      try await HealthClient(transport: stub).check(address: "http://mac.local")
    }
  }
}

private struct OfflineTransport: HealthTransport {
  func load(_ request: URLRequest) async throws -> (Data, URLResponse) {
    throw URLError(.cannotConnectToHost)
  }
}

@Test func reportsOfflineServer() async {
  await #expect(throws: HealthError.unavailable) {
    try await HealthClient(transport: OfflineTransport()).check(address: "http://mac.local")
  }
}

@Test func postsDebugLogToFixedPath() async throws {
  let transport = StubTransport()
  await DebugLog(transport: transport).report(
    address: "http://mac.local:8765", level: "info", message: "check started")
  let request = await transport.request
  #expect(request?.url?.absoluteString == "http://mac.local:8765/debug/log")
  #expect(request?.httpMethod == "POST")
  let event = try JSONDecoder().decode(DebugEvent.self, from: request?.httpBody ?? Data())
  #expect(event == DebugEvent(source: "tvos", level: "info", message: "check started"))
}

@Test func loadsLibraryFromEndpoint() async throws {
  let transport = StubTransport(
    body:
      "{\"items\":[{\"id\":\"abc\",\"title\":\"Heat\",\"filename\":\"Heat.mkv\",\"artwork\":\"/library/abc/artwork\"}]}"
  )
  let library = try await HealthClient(transport: transport).library(address: "http://mac.local")
  #expect(library.items.count == 1)
  #expect(library.items[0].title == "Heat")
  #expect(library.items[0].artwork == "/library/abc/artwork")
  let request = await transport.request
  #expect(request?.url?.absoluteString == "http://mac.local/library")
  #expect(request?.httpMethod == "GET")
}
