import HealthClient
import SwiftUI

struct ConnectionView: View {
  @AppStorage("serverAddress") private var address = ""
  @State private var state: ConnectionState = .ready
  @State private var checkTask: Task<Void, Never>?

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
      VStack(alignment: .leading, spacing: 30) {
        Text("NESTPLAY").font(.system(size: 24, weight: .semibold)).tracking(5).foregroundStyle(
          .secondary)
        Text("Your library starts here.").font(.system(size: 56, weight: .semibold))
        Text("Start the Nestplay server on your Mac, then enter its local address.")
          .font(.title3).foregroundStyle(.secondary)
        TextField("http://your-mac.local:8765", text: $address)
          .keyboardType(.URL)
          .textInputAutocapitalization(.never)
          .autocorrectionDisabled()
          .accessibilityLabel("Server address")
          .accessibilityIdentifier("serverAddress")
          .disabled(checking)
          .onSubmit { check() }
          .onChange(of: address) { if !checking { state = .ready } }
        Button(checking ? "Checking…" : "Check connection") { check() }
          .disabled(checking || address.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty)
          .accessibilityIdentifier("checkConnection")
        VStack(alignment: .leading, spacing: 12) {
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
        Text("Both devices must be on the same local network.")
          .font(.system(size: 20)).foregroundStyle(.secondary)
      }
      .frame(width: 1000, alignment: .leading)
      .padding(60)
    }
    .preferredColorScheme(.dark)
    .task {
      // Launch arguments allow the integration test to use the real local server.
      if let index = ProcessInfo.processInfo.arguments.firstIndex(of: "--server-url"),
        ProcessInfo.processInfo.arguments.indices.contains(index + 1)
      {
        address = ProcessInfo.processInfo.arguments[index + 1]
        check()
      }
    }
    .onDisappear {
      checkTask?.cancel()
      if checking { state = .ready }
    }
  }

  private var detail: String {
    switch state {
    case .ready: "The first check confirms your Apple TV can reach the server."
    case .checking: "Contacting your Mac…"
    case .connected: "Your Mac is reachable and ready. Media features will follow."
    case .failed(let message): message
    }
  }
  private var statusIcon: String {
    switch state {
    case .connected: "checkmark.circle.fill"
    case .failed: "wifi.exclamationmark"
    default: "desktopcomputer"
    }
  }
  private var statusColor: Color {
    switch state {
    case .connected: .green
    case .failed: .orange
    default: .secondary
    }
  }
  private func check() {
    guard !checking else { return }
    checkTask?.cancel()
    let submitted = address
    state = .checking
    checkTask = Task { @MainActor in
      do {
        _ = try await HealthClient().check(address: submitted)
        guard !Task.isCancelled else { return }
        state = .connected
      } catch {
        guard !Task.isCancelled else { return }
        state = .failed(error.localizedDescription)
      }
    }
  }
}
