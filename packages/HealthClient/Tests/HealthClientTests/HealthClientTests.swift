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

@Test func detectsLibraryChangedEvent() {
  #expect(libraryEventSignalsChange("data: {\"changed\":true}"))
  #expect(libraryEventSignalsChange("data: {\"changed\": true}"))
  #expect(!libraryEventSignalsChange(": ping"))
  #expect(!libraryEventSignalsChange(": connected"))
  #expect(!libraryEventSignalsChange("data: {\"changed\":false}"))
  #expect(!libraryEventSignalsChange("event: changed"))
}

@Test func buildsLibraryMediaURL() throws {
  let url = try HealthClient().mediaURL(address: "http://mac.local:8765", id: "abc")
  #expect(url.absoluteString == "http://mac.local:8765/library/abc/media")
}

@Test func loadsLibraryFromEndpoint() async throws {
  let transport = StubTransport(
    body:
      "{\"items\":[{\"id\":\"abc\",\"title\":\"Heat\",\"filename\":\"Heat.mkv\",\"collection\":\"\",\"artwork\":\"/library/abc/artwork\"}]}"
  )
  let library = try await HealthClient(transport: transport).library(address: "http://mac.local")
  #expect(library.items.count == 1)
  #expect(library.items[0].title == "Heat")
  #expect(library.items[0].collection == "")
  #expect(library.items[0].artwork == "/library/abc/artwork")
  let request = await transport.request
  #expect(request?.url?.absoluteString == "http://mac.local/library")
  #expect(request?.httpMethod == "GET")
}

@Test func ordersPlaybackQueueByFilenameInTheSameCollection() {
  let items = [
    CatalogItem(
      id: "11", title: "11", filename: "11 - B.mp4", artwork: "/a", collection: "Mateo/PJ Masks"),
    CatalogItem(
      id: "09", title: "09", filename: "09 - A.mp4", artwork: "/a", collection: "Mateo/PJ Masks"),
    CatalogItem(
      id: "simon", title: "Simon", filename: "02 - C.mp4", artwork: "/a",
      collection: "Mateo/Simon Castellano"),
  ]
  #expect(libraryPlaybackQueue(items: items, playing: "11").map(\.id) == ["09", "11"])
  #expect(libraryPlaybackQueue(items: items, playing: "missing").isEmpty)
}

@Test func browsesWatchedFolderTreeFromCollection() {
  let items = [
    CatalogItem(
      id: "09", title: "09", filename: "09.mp4", artwork: "/library/09/artwork",
      collection: "Mateo/PJ Masks"),
    CatalogItem(
      id: "simon", title: "Simon", filename: "02.mp4", artwork: "/library/simon/artwork",
      collection: "Mateo/Simon Castellano"),
    CatalogItem(
      id: "lim", title: "Limitless", filename: "E01.mkv", artwork: "/library/lim/artwork",
      collection: "Limitless with Chris Hemsworth S01"),
    CatalogItem(id: "root", title: "Loose", filename: "Loose.mp4", artwork: "/a", collection: ""),
  ]
  let root = libraryBrowse(items: items, path: "")
  #expect(root.folders.map(\.name) == ["Limitless with Chris Hemsworth S01", "Mateo"])
  #expect(root.folders.map(\.artwork) == ["/library/lim/artwork", "/library/simon/artwork"])
  #expect(root.files.map(\.id) == ["root"])
  let mateo = libraryBrowse(items: items, path: "Mateo")
  #expect(mateo.folders.map(\.name) == ["PJ Masks", "Simon Castellano"])
  #expect(mateo.folders.map(\.artwork) == ["/library/09/artwork", "/library/simon/artwork"])
  #expect(mateo.files.isEmpty)
  let pj = libraryBrowse(items: items, path: "Mateo/PJ Masks")
  #expect(pj.folders.isEmpty)
  #expect(pj.files.map(\.id) == ["09"])
}

@Test func loadsYouTubeHomeFromEndpoint() async throws {
  let transport = StubTransport(
    body:
      "{\"state\":\"connected\",\"items\":[{\"id\":\"abcdefghijk\",\"title\":\"Heat\",\"artwork\":\"/youtube/abcdefghijk/artwork\"}],\"cursor\":\"next\"}"
  )
  let page = try await HealthClient(transport: transport).youtubeHome(address: "http://mac.local")
  #expect(page.state == "connected")
  #expect(page.items.count == 1)
  #expect(page.items[0].title == "Heat")
  #expect(page.cursor == "next")
  let request = await transport.request
  #expect(request?.url?.absoluteString == "http://mac.local/youtube/home")
}

@Test func refreshesYouTubeHomeWithQuery() async throws {
  let transport = StubTransport(
    body: "{\"state\":\"connected\",\"items\":[],\"cursor\":\"\"}"
  )
  _ = try await HealthClient(transport: transport).youtubeHome(
    address: "http://mac.local", refresh: true)
  let request = await transport.request
  #expect(request?.url?.path == "/youtube/home")
  #expect(request?.url?.query?.contains("refresh=1") == true)
}

@Test func searchesYouTubeWithQuery() async throws {
  let transport = StubTransport(
    body: "{\"state\":\"connected\",\"items\":[],\"cursor\":\"\"}"
  )
  _ = try await HealthClient(transport: transport).youtubeSearch(
    address: "http://mac.local", query: "heat")
  let request = await transport.request
  #expect(request?.url?.path == "/youtube/search")
  #expect(request?.url?.query?.contains("q=heat") == true)
}

@Test func reconnectsYouTubeWithoutInventingHome() async throws {
  let transport = StubTransport(
    body:
      "{\"state\":\"reconnect_required\",\"message\":\"YouTube needs a reconnect on your Mac.\",\"items\":[]}"
  )
  let page = try await HealthClient(transport: transport).youtubeStatus(address: "http://mac.local")
  #expect(page.state == "reconnect_required")
  #expect(page.items.isEmpty)
}

@Test func enqueuesYouTubeItemsOnQueueEndpoint() async throws {
  let transport = StubTransport(
    body:
      "{\"items\":[{\"id\":\"dQw4w9WgXcQ\",\"title\":\"Heat\",\"state\":\"queued\"}]}"
  )
  let queued = try await HealthClient(transport: transport).enqueue(
    address: "http://mac.local",
    items: [YouTubeItem(id: "dQw4w9WgXcQ", title: "Heat", artwork: "/youtube/dQw4w9WgXcQ/artwork")]
  )
  #expect(queued.items.count == 1)
  #expect(queued.items[0].id == "dQw4w9WgXcQ")
  let request = await transport.request
  #expect(request?.url?.absoluteString == "http://mac.local/queue")
  #expect(request?.httpMethod == "POST")
}
