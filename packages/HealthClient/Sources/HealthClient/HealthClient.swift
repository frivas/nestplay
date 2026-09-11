import Foundation

public struct Health: Decodable, Sendable, Equatable {
  public let service: String
  public let status: String
  public let apiVersion: Int
  enum CodingKeys: String, CodingKey {
    case service, status
    case apiVersion = "api_version"
  }
}

public enum HealthError: LocalizedError, Equatable, Sendable {
  case invalidAddress, unavailable, invalidResponse, incompatibleServer

  public var errorDescription: String? {
    switch self {
    case .invalidAddress: "Enter the Mac's address, such as http://your-mac.local:8765."
    case .unavailable:
      "Couldn't reach your Mac. Check that the server is running and both devices are on the same network."
    case .invalidResponse:
      "The server answered, but didn't return a valid Nestplay health response."
    case .incompatibleServer: "This address isn't running a compatible Nestplay server."
    }
  }
}

public protocol HealthTransport: Sendable {
  func load(_ request: URLRequest) async throws -> (Data, URLResponse)
}

public struct SessionTransport: HealthTransport {
  private let session: URLSession
  public init() {
    let config = URLSessionConfiguration.ephemeral
    config.timeoutIntervalForRequest = 20
    config.timeoutIntervalForResource = 30
    config.urlCache = nil
    config.waitsForConnectivity = false
    session = URLSession(configuration: config)
  }
  public func load(_ request: URLRequest) async throws -> (Data, URLResponse) {
    try await session.data(for: request)
  }
}

public struct HealthClient: Sendable {
  private let transport: any HealthTransport
  public init(transport: any HealthTransport = SessionTransport()) { self.transport = transport }

  public func check(address: String) async throws -> Health {
    let endpoint = try nestplayEndpoint(address: address, path: "/health")
    var request = URLRequest(url: endpoint)
    request.httpMethod = "GET"
    request.cachePolicy = .reloadIgnoringLocalCacheData
    request.timeoutInterval = 5
    request.setValue("application/json", forHTTPHeaderField: "Accept")
    let data: Data
    let response: URLResponse
    do { (data, response) = try await transport.load(request) } catch is CancellationError {
      throw CancellationError()
    } catch { throw HealthError.unavailable }
    guard let http = response as? HTTPURLResponse, http.statusCode == 200,
      let health = try? JSONDecoder().decode(Health.self, from: data)
    else {
      throw HealthError.invalidResponse
    }
    guard health.service == "nestplay", health.status == "ok", health.apiVersion == 1 else {
      throw HealthError.incompatibleServer
    }
    return health
  }

  public func library(address: String) async throws -> Catalog {
    let endpoint = try nestplayEndpoint(address: address, path: "/library")
    var request = URLRequest(url: endpoint)
    request.httpMethod = "GET"
    request.cachePolicy = .reloadIgnoringLocalCacheData
    request.timeoutInterval = 5
    request.setValue("application/json", forHTTPHeaderField: "Accept")
    let data: Data
    let response: URLResponse
    do { (data, response) = try await transport.load(request) } catch is CancellationError {
      throw CancellationError()
    } catch { throw HealthError.unavailable }
    guard let http = response as? HTTPURLResponse, http.statusCode == 200,
      let library = try? JSONDecoder().decode(Catalog.self, from: data)
    else {
      throw HealthError.invalidResponse
    }
    return library
  }

  public func youtubeStatus(address: String) async throws -> YouTubePage {
    try await youtube(address: address, path: "/youtube/status")
  }

  public func youtubeHome(address: String, cursor: String = "") async throws -> YouTubePage {
    var query: [String: String] = [:]
    if !cursor.isEmpty { query["cursor"] = cursor }
    return try await youtube(address: address, path: "/youtube/home", query: query)
  }

  public func youtubeSearch(address: String, query: String, cursor: String = "") async throws
    -> YouTubePage
  {
    var fields = ["q": query]
    if !cursor.isEmpty { fields["cursor"] = cursor }
    return try await youtube(address: address, path: "/youtube/search", query: fields)
  }

  private func youtube(address: String, path: String, query: [String: String] = [:]) async throws
    -> YouTubePage
  {
    let endpoint = try nestplayEndpoint(address: address, path: path, query: query)
    var request = URLRequest(url: endpoint)
    request.httpMethod = "GET"
    request.cachePolicy = .reloadIgnoringLocalCacheData
    request.timeoutInterval = 20
    request.setValue("application/json", forHTTPHeaderField: "Accept")
    let data: Data
    let response: URLResponse
    do { (data, response) = try await transport.load(request) } catch is CancellationError {
      throw CancellationError()
    } catch { throw HealthError.unavailable }
    guard let http = response as? HTTPURLResponse, http.statusCode == 200,
      let page = try? JSONDecoder().decode(YouTubePage.self, from: data)
    else {
      throw HealthError.invalidResponse
    }
    return page
  }
}

public struct Catalog: Decodable, Sendable, Equatable {
  public let items: [CatalogItem]
}

public struct CatalogItem: Decodable, Sendable, Equatable, Identifiable {
  public let id: String
  public let title: String
  public let filename: String
  public let artwork: String
}

public struct YouTubePage: Decodable, Sendable, Equatable {
  public let state: String
  public let message: String?
  public let items: [YouTubeItem]
  public let cursor: String?
  public init(
    state: String, message: String? = nil, items: [YouTubeItem] = [], cursor: String? = nil
  ) {
    self.state = state
    self.message = message
    self.items = items
    self.cursor = cursor
  }
}

public struct YouTubeItem: Decodable, Sendable, Equatable, Identifiable {
  public let id: String
  public let title: String
  public let artwork: String
}

func nestplayEndpoint(address: String, path: String, query: [String: String] = [:]) throws -> URL {
  guard var url = URLComponents(string: address.trimmingCharacters(in: .whitespacesAndNewlines)),
    ["http", "https"].contains(url.scheme?.lowercased() ?? ""),
    let host = url.host, !host.isEmpty,
    url.user == nil, url.password == nil, url.query == nil, url.fragment == nil,
    url.path.isEmpty || url.path == "/"
  else { throw HealthError.invalidAddress }
  url.path = path
  if !query.isEmpty {
    url.queryItems = query.map { URLQueryItem(name: $0.key, value: $0.value) }
  }
  guard let endpoint = url.url else { throw HealthError.invalidAddress }
  return endpoint
}
