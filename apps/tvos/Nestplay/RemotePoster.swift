import SwiftUI

enum NestEgg {
  static let stage = Color(red: 0.043, green: 0.118, blue: 0.133)
  static let cream = Color(red: 0.965, green: 0.933, blue: 0.847)
  static let gold = Color(red: 0.910, green: 0.784, blue: 0.353)
  static let teal = Color(red: 0.106, green: 0.255, blue: 0.282)
}

struct RemotePoster: View {
  let title: String
  let artwork: String
  let server: String

  var body: some View {
    Button(action: {}) {
      VStack(alignment: .leading, spacing: 14) {
        poster
          .frame(maxWidth: .infinity)
          .aspectRatio(16 / 9, contentMode: .fit)
          .clipped()
        Text(title)
          .font(.headline)
          .foregroundStyle(NestEgg.cream)
          .lineLimit(2)
          .multilineTextAlignment(.leading)
          .frame(maxWidth: .infinity, alignment: .leading)
      }
    }
    .buttonStyle(.card)
    .tint(NestEgg.gold)
    .accessibilityLabel(title)
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
        colors: [NestEgg.teal, NestEgg.stage],
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
    let path = artwork.hasPrefix("/") ? artwork : "/" + artwork
    return URL(string: base + path)
  }
}
