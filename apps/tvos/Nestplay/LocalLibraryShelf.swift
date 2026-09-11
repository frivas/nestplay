import HealthClient
import SwiftUI

struct LocalLibraryShelf: View {
  let items: [CatalogItem]
  let server: String
  let emptyMessage: String

  private let columns = [
    GridItem(.flexible(), spacing: 36),
    GridItem(.flexible(), spacing: 36),
    GridItem(.flexible(), spacing: 36),
    GridItem(.flexible(), spacing: 36),
  ]

  var body: some View {
    VStack(alignment: .leading, spacing: 28) {
      Text("Local library")
        .font(.system(size: 44, weight: .semibold))
        .foregroundStyle(NestEgg.cream)
      if !emptyMessage.isEmpty {
        Text(emptyMessage)
          .font(.title3)
          .foregroundStyle(NestEgg.cream.opacity(0.7))
          .accessibilityIdentifier("libraryEmpty")
      }
      if !items.isEmpty {
        ScrollView {
          LazyVGrid(columns: columns, spacing: 44) {
            ForEach(items) { item in
              RemotePoster(title: item.title, artwork: item.artwork, server: server)
            }
          }
          .padding(.bottom, 80)
        }
        .accessibilityIdentifier("libraryCatalog")
      }
    }
    .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
  }
}
