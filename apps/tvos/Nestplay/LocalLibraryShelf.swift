import HealthClient
import SwiftUI

struct LocalLibraryShelf: View {
  let items: [CatalogItem]
  let server: String
  let emptyMessage: String

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
        PosterGrid(
          items: items,
          server: server,
          catalogID: "libraryCatalog",
          title: { $0.title },
          artwork: { $0.artwork }
        )
      }
    }
    .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
  }
}
