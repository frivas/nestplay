import HealthClient
import SwiftUI

struct LocalLibraryShelf: View {
  let items: [CatalogItem]
  let server: String
  let emptyMessage: String
  @Binding var folder: [String]

  private var path: String { folder.joined(separator: "/") }
  private var browse: LibraryBrowse { libraryBrowse(items: items, path: path) }
  private var title: String { folder.last ?? "Local library" }

  var body: some View {
    ScrollView {
      VStack(alignment: .leading, spacing: 28) {
        Text(title)
          .font(.system(size: 44, weight: .semibold))
          .foregroundStyle(NestEgg.cream)
        if browse.folders.isEmpty && browse.files.isEmpty {
          Text(emptyMessage.isEmpty ? "No videos in this folder yet." : emptyMessage)
            .font(.title3)
            .foregroundStyle(NestEgg.cream.opacity(0.7))
            .accessibilityIdentifier("libraryEmpty")
        }
        if !browse.folders.isEmpty {
          LazyVGrid(columns: columns, spacing: 56) {
            ForEach(browse.folders) { child in
              FolderPoster(title: child.name, artwork: child.artwork, server: server) {
                folder.append(child.name)
              }
            }
          }
          .padding(.horizontal, 24)
          .padding(.vertical, 8)
        }
        if !browse.files.isEmpty {
          PosterGrid(
            items: browse.files,
            server: server,
            catalogID: "libraryCatalog",
            title: { $0.title },
            artwork: { $0.artwork },
            onSelect: play,
            scrolls: false
          ) {
            EmptyView()
          }
        }
      }
    }
    .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
  }

  private let columns = [
    GridItem(.flexible(), spacing: 80),
    GridItem(.flexible(), spacing: 80),
    GridItem(.flexible(), spacing: 80),
  ]

  private func play(_ item: CatalogItem) {
    let ordered = libraryPlaybackQueue(items: items, playing: item.id)
    let client = HealthClient()
    var urls: [URL] = []
    urls.reserveCapacity(ordered.count)
    for episode in ordered {
      guard let url = try? client.mediaURL(address: server, id: episode.id) else { return }
      urls.append(url)
    }
    guard let start = ordered.firstIndex(where: { $0.id == item.id }) else { return }
    LibraryPlayer.shared.present(urls: urls, start: start)
  }
}
