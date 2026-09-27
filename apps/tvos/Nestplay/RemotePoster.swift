import SwiftUI

enum NestEgg {
  static let stage = Color(red: 0.043, green: 0.118, blue: 0.133)
  static let cream = Color(red: 0.965, green: 0.933, blue: 0.847)
  static let gold = Color(red: 0.910, green: 0.784, blue: 0.353)
  static let teal = Color(red: 0.106, green: 0.255, blue: 0.282)
  static let mist = Color(red: 0.62, green: 0.78, blue: 0.80)
  static let ember = Color(red: 0.86, green: 0.52, blue: 0.28)
}

enum NestChromeKind {
  case refresh
  case search
  case close
  case find
  case more

  var focusedFill: Color {
    switch self {
    case .refresh: NestEgg.gold
    case .search: NestEgg.cream
    case .close: NestEgg.mist
    case .find: NestEgg.ember
    case .more: NestEgg.gold
    }
  }

  var focusedLabel: Color { NestEgg.stage }
}

struct NestChromeButton: View {
  let title: String
  let kind: NestChromeKind
  var enabled: Bool = true
  var accessibilityID: String = ""
  let action: () -> Void

  var body: some View {
    Button(action: action) {
      Text(title)
        .font(.headline)
        .padding(.horizontal, 28)
        .padding(.vertical, 14)
    }
    .buttonStyle(NestChromeButtonStyle(kind: kind))
    .disabled(!enabled)
    .opacity(enabled ? 1 : 0.45)
    .focusEffectDisabled()
    .accessibilityIdentifier(accessibilityID)
  }
}

private struct NestChromeButtonStyle: ButtonStyle {
  let kind: NestChromeKind

  func makeBody(configuration: Configuration) -> some View {
    NestChromeButtonBody(configuration: configuration, kind: kind)
  }
}

private struct NestChromeButtonBody: View {
  let configuration: ButtonStyle.Configuration
  let kind: NestChromeKind
  @Environment(\.isFocused) private var isFocused

  var body: some View {
    configuration.label
      .foregroundStyle(isFocused ? kind.focusedLabel : NestEgg.cream)
      .background(Capsule().fill(isFocused ? kind.focusedFill : NestEgg.teal))
      .scaleEffect(isFocused ? 1.06 : 1)
      .animation(.easeOut(duration: 0.12), value: isFocused)
  }
}

struct FolderPoster: View {
  let title: String
  let artwork: String
  let server: String
  let onSelect: () -> Void

  var body: some View {
    Button(action: onSelect) {
      RemoteStill(artwork: artwork, server: server, placeholderSymbol: "folder.fill")
        .frame(maxWidth: .infinity)
        .aspectRatio(16 / 9, contentMode: .fit)
        .overlay {
          LinearGradient(
            stops: [
              .init(color: .clear, location: 0.42),
              .init(color: NestEgg.stage.opacity(0.88), location: 1),
            ],
            startPoint: .top,
            endPoint: .bottom
          )
        }
        .overlay(alignment: .topLeading) {
          Image(systemName: "folder.fill")
            .font(.system(size: 18, weight: .bold))
            .foregroundStyle(NestEgg.gold)
            .padding(10)
            .background(NestEgg.stage.opacity(0.62), in: Circle())
            .padding(.leading, 28)
            .padding(.top, 20)
        }
        .overlay(alignment: .bottomLeading) {
          Text(title)
            .font(.headline)
            .foregroundStyle(NestEgg.cream)
            .lineLimit(2)
            .multilineTextAlignment(.leading)
            .shadow(color: .black.opacity(0.55), radius: 6, y: 1)
            .padding(.horizontal, 28)
            .padding(.bottom, 22)
            .frame(maxWidth: .infinity, alignment: .leading)
        }
        .clipped()
    }
    .buttonStyle(.card)
    .tint(NestEgg.gold)
    .accessibilityLabel(title)
  }
}

struct RemotePoster: View {
  let title: String
  let artwork: String
  let server: String
  var onMoveUp: (() -> Void)? = nil
  var onSelect: (() -> Void)? = nil

  var body: some View {
    Button(action: { onSelect?() }) {
      VStack(alignment: .leading, spacing: 14) {
        RemoteStill(artwork: artwork, server: server)
          .frame(maxWidth: .infinity)
          .aspectRatio(16 / 9, contentMode: .fit)
          .clipped()
        Text(title)
          .font(.headline)
          .foregroundStyle(NestEgg.cream)
          .lineLimit(2)
          .multilineTextAlignment(.leading)
          .frame(maxWidth: .infinity, alignment: .leading)
          .padding(.horizontal, 8)
      }
    }
    .buttonStyle(.card)
    .tint(NestEgg.gold)
    .onMoveCommand { direction in
      guard direction == .up else { return }
      onMoveUp?()
    }
    .accessibilityLabel(title)
  }

}

struct RemoteStill: View {
  let artwork: String
  let server: String
  var placeholderSymbol: String = "play.fill"

  var body: some View {
    still
      .frame(maxWidth: .infinity, maxHeight: .infinity)
      .clipped()
  }

  @ViewBuilder private var still: some View {
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
      .transaction { $0.animation = nil }
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
      Image(systemName: placeholderSymbol)
        .font(.system(size: 36, weight: .semibold))
        .foregroundStyle(placeholderSymbol == "folder.fill" ? NestEgg.gold : NestEgg.cream)
    }
  }

  private var artworkURL: URL? {
    guard !artwork.isEmpty else { return nil }
    let base = server.trimmingCharacters(in: CharacterSet(charactersIn: "/"))
    let path = artwork.hasPrefix("/") ? artwork : "/" + artwork
    return URL(string: base + path)
  }
}

struct PosterGrid<Item: Identifiable, Footer: View>: View {
  let items: [Item]
  let server: String
  let catalogID: String
  let title: (Item) -> String
  let artwork: (Item) -> String
  var footer: Footer
  var onMoveUpFromTopRow: (() -> Void)?
  var onSelect: ((Item) -> Void)?
  var scrolls: Bool = true

  init(
    items: [Item],
    server: String,
    catalogID: String,
    title: @escaping (Item) -> String,
    artwork: @escaping (Item) -> String,
    onMoveUpFromTopRow: (() -> Void)? = nil,
    onSelect: ((Item) -> Void)? = nil,
    scrolls: Bool = true,
    @ViewBuilder footer: () -> Footer
  ) {
    self.items = items
    self.server = server
    self.catalogID = catalogID
    self.title = title
    self.artwork = artwork
    self.onMoveUpFromTopRow = onMoveUpFromTopRow
    self.onSelect = onSelect
    self.scrolls = scrolls
    self.footer = footer()
  }

  private let columns = [
    GridItem(.flexible(), spacing: 80),
    GridItem(.flexible(), spacing: 80),
    GridItem(.flexible(), spacing: 80),
  ]

  var body: some View {
    Group {
      if scrolls {
        ScrollView { grid }
      } else {
        grid
      }
    }
    .focusSection()
    .accessibilityIdentifier(catalogID)
  }

  private var grid: some View {
    VStack(alignment: .leading, spacing: 28) {
      LazyVGrid(columns: columns, spacing: 56) {
        ForEach(Array(items.enumerated()), id: \.element.id) { index, item in
          RemotePoster(
            title: title(item),
            artwork: artwork(item),
            server: server,
            onMoveUp: index < columns.count ? onMoveUpFromTopRow : nil,
            onSelect: { onSelect?(item) }
          )
        }
      }
      footer
    }
    .padding(.horizontal, 24)
    .padding(.vertical, 28)
    .padding(.bottom, 80)
  }
}
