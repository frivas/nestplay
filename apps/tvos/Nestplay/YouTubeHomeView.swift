import HealthClient
import SwiftUI

struct YouTubeHomeView: View {
  let server: String
  @State private var page = YouTubePage(state: "checking", message: nil, items: [], cursor: nil)
  @State private var searching = false
  @State private var query = ""
  @State private var loading = false

  private let columns = [
    GridItem(.flexible(), spacing: 36),
    GridItem(.flexible(), spacing: 36),
    GridItem(.flexible(), spacing: 36),
    GridItem(.flexible(), spacing: 36),
  ]

  var body: some View {
    VStack(alignment: .leading, spacing: 24) {
      header
      if searching {
        searchField
      }
      if loading && page.items.isEmpty {
        ProgressView().tint(NestEgg.cream)
      }
      if let message = page.message, page.items.isEmpty {
        Text(message)
          .font(.title3)
          .foregroundStyle(NestEgg.cream.opacity(0.7))
          .accessibilityIdentifier("youtubeHomeEmpty")
      } else if page.items.isEmpty && !loading {
        Text(searching ? "No landscape search results." : "No landscape recommendations right now.")
          .font(.title3)
          .foregroundStyle(NestEgg.cream.opacity(0.7))
          .accessibilityIdentifier("youtubeHomeEmpty")
      }
      if !page.items.isEmpty {
        ScrollView {
          LazyVGrid(columns: columns, spacing: 44) {
            ForEach(page.items) { item in
              RemotePoster(title: item.title, artwork: item.artwork, server: server)
            }
          }
          .padding(NestEgg.focusGutter)
          if let cursor = page.cursor, !cursor.isEmpty {
            Button(searching ? "More results" : "More recommendations") {
              Task { await loadMore() }
            }
            .buttonStyle(.borderless)
            .tint(NestEgg.gold)
            .padding(.bottom, 80)
          }
        }
        .scrollClipDisabled()
        .accessibilityIdentifier("youtubeHomeCatalog")
      }
    }
    .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
    .accessibilityIdentifier("youtubeHomeSurface")
    .task { await refresh() }
  }

  private var header: some View {
    HStack(alignment: .center, spacing: 24) {
      Text(searching ? "YouTube search" : "YouTube Home")
        .font(.system(size: 44, weight: .semibold))
        .foregroundStyle(NestEgg.cream)
      Spacer()
      if searching {
        Button("Close search") {
          searching = false
          query = ""
          Task { await refresh() }
        }
        .buttonStyle(.borderless)
        .tint(NestEgg.gold)
      } else {
        Button("Refresh") { Task { await refresh() } }
          .disabled(loading)
          .buttonStyle(.borderless)
          .tint(NestEgg.gold)
          .accessibilityIdentifier("youtubeHomeRefresh")
        Button("Search") { searching = true }
          .disabled(page.state == "reconnect_required")
          .buttonStyle(.borderless)
          .tint(NestEgg.gold)
          .accessibilityIdentifier("youtubeHomeSearch")
      }
    }
  }

  private var searchField: some View {
    HStack(spacing: 24) {
      TextField("Search YouTube", text: $query)
        .textFieldStyle(.plain)
        .font(.title2)
        .foregroundStyle(NestEgg.cream)
        .accessibilityIdentifier("youtubeSearchField")
      Button("Find") { Task { await search() } }
        .disabled(query.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty || loading)
        .buttonStyle(.borderless)
        .tint(NestEgg.gold)
        .accessibilityIdentifier("youtubeSearchSubmit")
    }
  }

  private func refresh() async {
    loading = true
    defer { loading = false }
    do {
      page = try await HealthClient().youtubeHome(address: server)
    } catch {
      page = YouTubePage(
        state: "unavailable",
        message: "YouTube Home could not be loaded.",
        items: [],
        cursor: nil)
    }
  }

  private func search() async {
    let q = query.trimmingCharacters(in: .whitespacesAndNewlines)
    guard !q.isEmpty else { return }
    loading = true
    defer { loading = false }
    do {
      page = try await HealthClient().youtubeSearch(address: server, query: q)
    } catch {
      page = YouTubePage(
        state: "unavailable",
        message: "YouTube search could not be loaded.",
        items: [],
        cursor: nil)
    }
  }

  private func loadMore() async {
    guard let cursor = page.cursor, !cursor.isEmpty else { return }
    loading = true
    defer { loading = false }
    do {
      let next: YouTubePage
      if searching {
        next = try await HealthClient().youtubeSearch(
          address: server, query: query, cursor: cursor)
      } else {
        next = try await HealthClient().youtubeHome(address: server, cursor: cursor)
      }
      page = YouTubePage(
        state: next.state,
        message: next.message,
        items: page.items + next.items,
        cursor: next.cursor)
    } catch {
      page = YouTubePage(
        state: page.state,
        message: "Could not load more videos.",
        items: page.items,
        cursor: nil)
    }
  }
}
