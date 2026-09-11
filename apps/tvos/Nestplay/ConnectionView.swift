import HealthClient
import OSLog
import SwiftUI

struct ConnectionView: View {
  @State private var address = PocServer.address
  @State private var state: ConnectionState = .ready
  @State private var catalog: [CatalogItem] = []
  @State private var catalogMessage = ""
  @State private var surface: Surface = .localLibrary
  @State private var checkTask: Task<Void, Never>?
  private let log = Logger(subsystem: "local.nestplay.poc", category: "connection")

  private enum Surface: Hashable {
    case localLibrary
    case youtubeHome
  }

  private enum ConnectionState {
    case ready, checking, connected
    case failed(String)
    var title: String {
      switch self {
      case .ready: "Connect to your Mac"
      case .checking: "Checking connection…"
      case .connected: "Connected to Nestplay"
      case .failed: "Connection needs attention"
      }
    }
  }

  private var checking: Bool { if case .checking = state { true } else { false } }

  var body: some View {
    ZStack {
      Color(red: 0.055, green: 0.07, blue: 0.09).ignoresSafeArea()
      if case .connected = state {
        connectedShell
      } else {
        setupShell
      }
    }
    .preferredColorScheme(.dark)
    .task {
      if let index = ProcessInfo.processInfo.arguments.firstIndex(of: "--server-url"),
        ProcessInfo.processInfo.arguments.indices.contains(index + 1)
      {
        address = ProcessInfo.processInfo.arguments[index + 1]
      } else {
        address = PocServer.address
      }
      trace("info", "launch address=\(address)")
      check()
    }
    .onDisappear {
      trace("debug", "view disappeared state=\(state.title)")
      checkTask?.cancel()
      if checking { state = .ready }
    }
  }

  private var setupShell: some View {
    VStack(alignment: .leading, spacing: 30) {
      Text("NESTPLAY").font(.system(size: 24, weight: .semibold)).tracking(5).foregroundStyle(
        .secondary)
      Text("Your library starts here.").font(.system(size: 56, weight: .semibold))
      Text("This PoC uses a fixed Mac address and checks it when the app opens.")
        .font(.title3).foregroundStyle(.secondary)
      Text(address)
        .font(.title3.monospaced())
        .foregroundStyle(.secondary)
        .accessibilityLabel("Server address")
        .accessibilityIdentifier("serverAddress")
      connectionControls
      Text("Traces from this Apple TV are printed by the Nestplay server.")
        .font(.system(size: 20)).foregroundStyle(.secondary)
    }
    .frame(width: 1000, alignment: .leading)
    .padding(60)
  }

  private var connectedShell: some View {
    TabView(selection: $surface) {
      localLibraryTab
        .tabItem {
          Text("Local library")
        }
        .tag(Surface.localLibrary)
        .accessibilityIdentifier("localLibraryTab")
      YouTubeHomeView(server: address)
        .tabItem {
          Text("YouTube Home")
        }
        .tag(Surface.youtubeHome)
        .accessibilityIdentifier("youtubeHomeTab")
    }
  }

  private var localLibraryTab: some View {
    VStack(alignment: .leading, spacing: 24) {
      HStack(alignment: .center, spacing: 24) {
        Text("NESTPLAY")
          .font(.system(size: 22, weight: .semibold))
          .tracking(6)
          .foregroundStyle(Color(red: 0.965, green: 0.933, blue: 0.847).opacity(0.72))
        Spacer()
        Text(state.title)
          .font(.headline)
          .foregroundStyle(Color(red: 0.965, green: 0.933, blue: 0.847).opacity(0.8))
          .accessibilityIdentifier("connectionStatus")
      }
      LocalLibraryShelf(items: catalog, server: address, emptyMessage: catalogMessage)
    }
    .padding(.horizontal, 72)
    .padding(.top, 36)
    .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
    .accessibilityIdentifier("localLibrarySurface")
  }

  private var connectionControls: some View {
    VStack(alignment: .leading, spacing: 12) {
      Button(checking ? "Checking…" : "Check connection") { check() }
        .disabled(checking)
        .accessibilityIdentifier("checkConnection")
      HStack(spacing: 16) {
        if checking {
          ProgressView()
        } else {
          Image(systemName: statusIcon).foregroundStyle(statusColor)
        }
        Text(state.title).font(.title3.weight(.medium)).accessibilityIdentifier(
          "connectionStatus")
      }
      Text(detail).font(.system(size: 23)).foregroundStyle(.secondary)
        .fixedSize(horizontal: false, vertical: true)
    }
    .frame(maxWidth: .infinity, minHeight: 110, alignment: .topLeading)
  }

  private var detail: String {
    switch state {
    case .ready: "The first check confirms your Apple TV can reach the server."
    case .checking: "Contacting your Mac…"
    case .connected: "Your Mac is reachable. The list below is read-only."
    case .failed(let message): message
    }
  }
  private var statusIcon: String {
    switch state {
    case .connected: "checkmark.circle.fill"
    case .failed: "wifi.exclamationmark"
    case .ready, .checking: "desktopcomputer"
    }
  }
  private var statusColor: Color {
    switch state {
    case .connected: .green
    case .failed: .orange
    case .ready, .checking: .secondary
    }
  }
  private func check() {
    guard !checking else { return }
    checkTask?.cancel()
    let submitted = address
    state = .checking
    surface = .localLibrary
    catalog = []
    catalogMessage = ""
    trace("info", "check started address=\(submitted)")
    checkTask = Task { @MainActor in
      do {
        _ = try await HealthClient().check(address: submitted)
        guard !Task.isCancelled else {
          trace("debug", "check cancelled address=\(submitted)")
          return
        }
        trace("info", "check connected address=\(submitted)")
        state = .connected
        do {
          let library = try await HealthClient().library(address: submitted)
          guard !Task.isCancelled else { return }
          catalog = library.items
          catalogMessage = library.items.isEmpty ? "No videos in the watched folder yet." : ""
          trace("info", "library loaded count=\(library.items.count)")
        } catch {
          guard !Task.isCancelled else { return }
          catalogMessage = "Connected, but the library list could not be loaded."
          trace("error", "library failed error=\(error.localizedDescription)")
        }
      } catch {
        guard !Task.isCancelled else {
          trace("debug", "check cancelled address=\(submitted)")
          return
        }
        trace("error", "check failed address=\(submitted) error=\(error.localizedDescription)")
        state = .failed(error.localizedDescription)
      }
    }
  }

  private func trace(_ level: String, _ message: String) {
    switch level {
    case "error": log.error("\(message, privacy: .public)")
    case "info": log.info("\(message, privacy: .public)")
    default: log.debug("\(message, privacy: .public)")
    }
    let address = address
    Task { await DebugLog().report(address: address, level: level, message: message) }
  }
}
