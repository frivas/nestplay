import HealthClient
import SwiftUI

private enum NestEgg {
  static let stage = Color(red: 0.043, green: 0.118, blue: 0.133)
  static let cream = Color(red: 0.965, green: 0.933, blue: 0.847)
  static let gold = Color(red: 0.910, green: 0.784, blue: 0.353)
}

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
              PosterCard(item: item, server: server)
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

private struct PosterCard: View {
  let item: CatalogItem
  let server: String

  var body: some View {
    Button(action: {}) {
      VStack(alignment: .leading, spacing: 14) {
        poster
          .frame(maxWidth: .infinity)
          .aspectRatio(16 / 9, contentMode: .fit)
          .clipped()
        Text(item.title)
          .font(.headline)
          .foregroundStyle(NestEgg.cream)
          .lineLimit(2)
          .multilineTextAlignment(.leading)
          .frame(maxWidth: .infinity, alignment: .leading)
      }
    }
    .buttonStyle(.card)
    .tint(NestEgg.gold)
    .accessibilityLabel(item.title)
  }

  @ViewBuilder private var poster: some View {
    if let url = artworkURL {
      AsyncImage(url: url) { phase in
        switch phase {
        case .success(let image):
          image.resizable().scaledToFill()
        case .failure:
          placeholder
        case .empty:
          ZStack {
            placeholder
            ProgressView().tint(NestEgg.cream)
          }
        @unknown default:
          placeholder
        }
      }
    } else {
      placeholder
    }
  }

  private var placeholder: some View {
    ZStack {
      LinearGradient(
        colors: [
          Color(red: 0.106, green: 0.255, blue: 0.282),
          NestEgg.stage,
        ],
        startPoint: .top,
        endPoint: .bottom
      )
      Image(systemName: "play.fill")
        .font(.system(size: 36, weight: .semibold))
        .foregroundStyle(NestEgg.cream)
    }
  }

  private var artworkURL: URL? {
    let base = server.trimmingCharacters(in: CharacterSet(charactersIn: "/"))
    let path = item.artwork.hasPrefix("/") ? item.artwork : "/" + item.artwork
    return URL(string: base + path)
  }
}
