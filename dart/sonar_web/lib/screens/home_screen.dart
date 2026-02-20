import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';

import '../providers/auth_provider.dart';
import '../services/media_service.dart';
import 'logister_modal.dart';

class HomeScreen extends StatelessWidget {
  const HomeScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Stack(
        children: [
          const _LandingBackground(),
          SafeArea(
            child: Consumer<AuthProvider>(
              builder: (context, auth, _) {
                if (auth.loading) {
                  return const Center(child: CircularProgressIndicator());
                }
                if (auth.isAuthenticated) {
                  return const SizedBox.shrink(); // redirect handles navigation
                }
                return _HomeContent();
              },
            ),
          ),
        ],
      ),
    );
  }
}

class _HomeContent extends StatefulWidget {
  @override
  State<_HomeContent> createState() => _HomeContentState();
}

class _HomeContentState extends State<_HomeContent> {
  bool _showLogister = false;

  @override
  Widget build(BuildContext context) {
    final from = Uri.base.queryParameters['from'];
    final shouldShowLogister = _showLogister || (from != null && from.isNotEmpty);

    if (shouldShowLogister) {
      final mediaService = context.read<MediaService>();
      return LogisterModal(
        mediaService: mediaService,
        onSuccess: () {
          final dest = from != null && from.isNotEmpty
              ? Uri.decodeComponent(from)
              : '/single-player';
          context.go(dest);
        },
        onSkip: () {
          if (from != null && from.isNotEmpty) {
            context.go('/');
          } else {
            setState(() => _showLogister = false);
          }
        },
      );
    }

    final colorScheme = Theme.of(context).colorScheme;

    return LayoutBuilder(
      builder: (context, constraints) {
        final isWide = constraints.maxWidth > 980;
        final sidePadding = isWide ? 48.0 : 24.0;
        final statWidth = isWide ? 260.0 : constraints.maxWidth;
        final cardWidth = isWide
            ? (constraints.maxWidth - 32) / 2
            : constraints.maxWidth;

        return Align(
          alignment: Alignment.topCenter,
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 1200),
            child: SingleChildScrollView(
              padding: EdgeInsets.fromLTRB(sidePadding, 32, sidePadding, 64),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  _AnimatedReveal(
                    delay: 0.0,
                    child: Text(
                      'Unclaimed Streets',
                      style: Theme.of(context).textTheme.titleMedium?.copyWith(
                            letterSpacing: 3,
                            fontWeight: FontWeight.w600,
                            color: colorScheme.primary,
                          ),
                    ),
                  ),
                  const SizedBox(height: 12),
                  _AnimatedReveal(
                    delay: 0.06,
                    child: Text(
                      'Step into the uncharted and map what the world hides.',
                      style: Theme.of(context).textTheme.displaySmall?.copyWith(
                            fontWeight: FontWeight.w600,
                            color: colorScheme.onBackground,
                          ),
                    ),
                  ),
                  const SizedBox(height: 16),
                  _AnimatedReveal(
                    delay: 0.12,
                    child: Text(
                      'Explore real landscapes, uncover mythic locations, and grow your strength with every quest. Unclaimed Streets guides your crew through discoveries that feel grounded, mysterious, and alive.',
                      style: Theme.of(context).textTheme.titleMedium?.copyWith(
                            color: colorScheme.onBackground.withOpacity(0.8),
                            height: 1.4,
                          ),
                    ),
                  ),
                  const SizedBox(height: 24),
                  _AnimatedReveal(
                    delay: 0.18,
                    child: Wrap(
                      spacing: 16,
                      runSpacing: 16,
                      children: [
                        SizedBox(
                          width: statWidth,
                          child: const _HeroStat(
                            icon: Icons.explore,
                            title: 'Worldbound exploration',
                            subtitle: 'Real places, deeper stories',
                            accent: Color(0xFF355C7D),
                          ),
                        ),
                        SizedBox(
                          width: statWidth,
                          child: const _HeroStat(
                            icon: Icons.auto_awesome,
                            title: 'Mythic overlays',
                            subtitle: 'Ley lines and hidden gates',
                            accent: Color(0xFFB87333),
                          ),
                        ),
                        SizedBox(
                          width: statWidth,
                          child: const _HeroStat(
                            icon: Icons.bolt,
                            title: 'Strength growth',
                            subtitle: 'Power up with every quest',
                            accent: Color(0xFF6B8E23),
                          ),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(height: 28),
                  _AnimatedReveal(
                    delay: 0.24,
                    child: isWide
                        ? Row(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Expanded(
                                flex: 3,
                                child: _ExpeditionCallout(
                                  onStart: () => setState(() => _showLogister = true),
                                ),
                              ),
                              const SizedBox(width: 24),
                              Expanded(
                                flex: 2,
                                child: _ExpeditionDossier(
                                  onStart: () => setState(() => _showLogister = true),
                                ),
                              ),
                            ],
                          )
                        : Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              _ExpeditionCallout(
                                onStart: () => setState(() => _showLogister = true),
                              ),
                              const SizedBox(height: 20),
                              _ExpeditionDossier(
                                onStart: () => setState(() => _showLogister = true),
                              ),
                            ],
                          ),
                  ),
                  const SizedBox(height: 32),
                  _AnimatedReveal(
                    delay: 0.3,
                    child: Text(
                      'Your quest path',
                      style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                            color: colorScheme.onBackground,
                            fontWeight: FontWeight.w600,
                          ),
                    ),
                  ),
                  const SizedBox(height: 12),
                  _AnimatedReveal(
                    delay: 0.34,
                    child: Text(
                      'Each mission blends real-world discovery with a layer of legend. Move through zones, claim artifacts, and complete quest chains to unlock the next realm.',
                      style: Theme.of(context).textTheme.titleMedium?.copyWith(
                            color: colorScheme.onBackground.withOpacity(0.75),
                          ),
                    ),
                  ),
                  const SizedBox(height: 20),
                  _AnimatedReveal(
                    delay: 0.38,
                    child: Wrap(
                      spacing: 16,
                      runSpacing: 16,
                      children: [
                        SizedBox(
                          width: cardWidth,
                          child: const _QuestCard(
                            icon: Icons.travel_explore,
                            title: 'Discover the living map',
                            description:
                                'Navigate real landmarks, then reveal the mystic layer only visible to your crew.',
                            accent: Color(0xFF355C7D),
                          ),
                        ),
                        SizedBox(
                          width: cardWidth,
                          child: const _QuestCard(
                            icon: Icons.auto_graph,
                            title: 'Grow your strength',
                            description:
                                'Complete quests to unlock new abilities, upgrades, and movement boosts.',
                            accent: Color(0xFF6B8E23),
                          ),
                        ),
                        SizedBox(
                          width: cardWidth,
                          child: const _QuestCard(
                            icon: Icons.lightbulb,
                            title: 'Unearth mythic sites',
                            description:
                                'Track anomalies, hidden ruins, and ley lines across the map.',
                            accent: Color(0xFFB87333),
                          ),
                        ),
                        SizedBox(
                          width: cardWidth,
                          child: const _QuestCard(
                            icon: Icons.emoji_events,
                            title: 'Complete quest chains',
                            description:
                                'Finish each chain to open the next realm and earn legendary rewards.',
                            accent: Color(0xFF0B5563),
                          ),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(height: 28),
                  _AnimatedReveal(
                    delay: 0.42,
                    child: _ExpeditionSteps(),
                  ),
                ],
              ),
            ),
          ),
        );
      },
    );
  }
}

class _AnimatedReveal extends StatelessWidget {
  const _AnimatedReveal({
    required this.child,
    this.delay = 0.0,
    this.offset = const Offset(0, 24),
    this.duration = const Duration(milliseconds: 900),
  });

  final Widget child;
  final double delay;
  final Offset offset;
  final Duration duration;

  @override
  Widget build(BuildContext context) {
    return TweenAnimationBuilder<double>(
      tween: Tween(begin: 0, end: 1),
      duration: duration,
      curve: Interval(delay, 1.0, curve: Curves.easeOutCubic),
      builder: (context, value, child) {
        return Opacity(
          opacity: value,
          child: Transform.translate(
            offset: Offset(offset.dx * (1 - value), offset.dy * (1 - value)),
            child: child,
          ),
        );
      },
      child: child,
    );
  }
}

class _LandingBackground extends StatelessWidget {
  const _LandingBackground();

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    return Stack(
      children: [
        Container(
          decoration: BoxDecoration(
            gradient: LinearGradient(
              colors: [
                const Color(0xFFFFF4E2),
                const Color(0xFFF6EAD2),
                const Color(0xFFE9D3AE),
              ],
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
            ),
          ),
        ),
        Positioned.fill(
          child: CustomPaint(
            painter: _TopoLinesPainter(
              colorScheme.onBackground.withOpacity(0.08),
            ),
          ),
        ),
        Positioned(
          top: -160,
          right: -120,
          child: _GlowOrb(
            size: 360,
            colors: [
              colorScheme.primary.withOpacity(0.25),
              Colors.transparent,
            ],
          ),
        ),
        Positioned(
          bottom: -180,
          left: -140,
          child: _GlowOrb(
            size: 380,
            colors: [
              colorScheme.tertiary.withOpacity(0.2),
              Colors.transparent,
            ],
          ),
        ),
        Positioned(
          top: 120,
          left: 40,
          child: _GlowOrb(
            size: 160,
            colors: [
              const Color(0xFF6B8E23).withOpacity(0.18),
              Colors.transparent,
            ],
          ),
        ),
      ],
    );
  }
}

class _GlowOrb extends StatelessWidget {
  const _GlowOrb({
    required this.size,
    required this.colors,
  });

  final double size;
  final List<Color> colors;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: size,
      height: size,
      decoration: BoxDecoration(
        shape: BoxShape.circle,
        gradient: RadialGradient(
          colors: colors,
        ),
      ),
    );
  }
}

class _HeroStat extends StatelessWidget {
  const _HeroStat({
    required this.icon,
    required this.title,
    required this.subtitle,
    required this.accent,
  });

  final IconData icon;
  final String title;
  final String subtitle;
  final Color accent;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: colorScheme.surface.withOpacity(0.92),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: accent.withOpacity(0.4)),
        boxShadow: [
          BoxShadow(
            color: accent.withOpacity(0.12),
            blurRadius: 18,
            offset: const Offset(0, 10),
          ),
        ],
      ),
      child: Row(
        children: [
          Container(
            width: 44,
            height: 44,
            decoration: BoxDecoration(
              color: accent.withOpacity(0.12),
              borderRadius: BorderRadius.circular(14),
              border: Border.all(color: accent.withOpacity(0.35)),
            ),
            child: Icon(icon, color: accent),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  title,
                  style: Theme.of(context).textTheme.titleMedium?.copyWith(
                        fontWeight: FontWeight.w600,
                        color: colorScheme.onSurface,
                      ),
                ),
                const SizedBox(height: 4),
                Text(
                  subtitle,
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(
                        color: colorScheme.onSurface.withOpacity(0.7),
                      ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _ExpeditionCallout extends StatelessWidget {
  const _ExpeditionCallout({required this.onStart});

  final VoidCallback onStart;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    return Container(
      padding: const EdgeInsets.all(24),
      decoration: BoxDecoration(
        color: colorScheme.surface.withOpacity(0.92),
        borderRadius: BorderRadius.circular(24),
        border: Border.all(color: colorScheme.outlineVariant),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.08),
            blurRadius: 24,
            offset: const Offset(0, 12),
          ),
        ],
      ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                'The expedition begins with your signal.',
                style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                      fontWeight: FontWeight.w600,
                      color: colorScheme.onSurface,
                ),
          ),
          const SizedBox(height: 12),
          Text(
            'Claim your route, connect with your crew, and set your sights on the next quest chain.',
            style: Theme.of(context).textTheme.titleMedium?.copyWith(
                  color: colorScheme.onSurface.withOpacity(0.75),
                ),
          ),
          const SizedBox(height: 20),
          Row(
            children: [
              FilledButton.icon(
                onPressed: onStart,
                icon: const Icon(Icons.wifi_tethering),
                label: const Text('Begin expedition'),
              ),
              const SizedBox(width: 12),
              OutlinedButton(
                onPressed: onStart,
                child: const Text('Join your crew'),
              ),
            ],
          ),
        ],
      ),
    );
  }
}

class _ExpeditionDossier extends StatelessWidget {
  const _ExpeditionDossier({required this.onStart});

  final VoidCallback onStart;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        gradient: LinearGradient(
          colors: [
            colorScheme.surfaceVariant.withOpacity(0.95),
            colorScheme.surface.withOpacity(0.95),
          ],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        borderRadius: BorderRadius.circular(24),
        border: Border.all(color: colorScheme.outlineVariant),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Expedition dossier',
            style: Theme.of(context).textTheme.titleLarge?.copyWith(
                  fontWeight: FontWeight.w600,
                ),
          ),
          const SizedBox(height: 8),
          Text(
            'Your crew is ready. Confirm your signal to unlock the world map, quest log, and strength track.',
            style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                  color: colorScheme.onSurface.withOpacity(0.72),
                ),
          ),
          const SizedBox(height: 16),
          _DossierItem(
            icon: Icons.map,
            title: 'World map',
            description: 'Live exploration with mythic overlays',
          ),
          const SizedBox(height: 12),
          _DossierItem(
            icon: Icons.shield,
            title: 'Strength track',
            description: 'Level up after every quest chain',
          ),
          const SizedBox(height: 12),
          _DossierItem(
            icon: Icons.task_alt,
            title: 'Quest log',
            description: 'Track real and mystical objectives',
          ),
          const SizedBox(height: 16),
          FilledButton(
            onPressed: onStart,
            child: const Text('Claim your signal'),
          ),
        ],
      ),
    );
  }
}

class _DossierItem extends StatelessWidget {
  const _DossierItem({
    required this.icon,
    required this.title,
    required this.description,
  });

  final IconData icon;
  final String title;
  final String description;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    return Row(
      children: [
        Container(
          width: 36,
          height: 36,
          decoration: BoxDecoration(
            color: colorScheme.primary.withOpacity(0.12),
            borderRadius: BorderRadius.circular(12),
          ),
          child: Icon(icon, color: colorScheme.primary, size: 18),
        ),
        const SizedBox(width: 12),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                title,
                style: Theme.of(context).textTheme.titleSmall?.copyWith(
                      fontWeight: FontWeight.w600,
                    ),
              ),
              Text(
                description,
                style: Theme.of(context).textTheme.bodySmall?.copyWith(
                      color: colorScheme.onSurface.withOpacity(0.7),
                    ),
              ),
            ],
          ),
        ),
      ],
    );
  }
}

class _QuestCard extends StatelessWidget {
  const _QuestCard({
    required this.icon,
    required this.title,
    required this.description,
    required this.accent,
  });

  final IconData icon;
  final String title;
  final String description;
  final Color accent;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: colorScheme.surface.withOpacity(0.95),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: accent.withOpacity(0.35)),
        boxShadow: [
          BoxShadow(
            color: accent.withOpacity(0.12),
            blurRadius: 20,
            offset: const Offset(0, 12),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            width: 48,
            height: 48,
            decoration: BoxDecoration(
              color: accent.withOpacity(0.12),
              borderRadius: BorderRadius.circular(14),
            ),
            child: Icon(icon, color: accent, size: 24),
          ),
          const SizedBox(height: 14),
          Text(
            title,
            style: Theme.of(context).textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.w600,
                  color: colorScheme.onSurface,
                ),
          ),
          const SizedBox(height: 8),
          Text(
            description,
            style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                  color: colorScheme.onSurface.withOpacity(0.75),
                ),
          ),
        ],
      ),
    );
  }
}

class _ExpeditionSteps extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: colorScheme.surface.withOpacity(0.92),
        borderRadius: BorderRadius.circular(22),
        border: Border.all(color: colorScheme.outlineVariant),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Launch sequence',
            style: Theme.of(context).textTheme.titleLarge?.copyWith(
                  fontWeight: FontWeight.w600,
                ),
          ),
          const SizedBox(height: 12),
          const _ExpeditionStep(
            index: '01',
            title: 'Send your signal',
            description: 'Request the access pulse for your phone number.',
          ),
          const SizedBox(height: 12),
          const _ExpeditionStep(
            index: '02',
            title: 'Confirm the gate',
            description: 'Verify the pulse and lock in your expedition crew.',
          ),
          const SizedBox(height: 12),
          const _ExpeditionStep(
            index: '03',
            title: 'Enter the world',
            description: 'Open the map, track quests, and begin exploring.',
          ),
        ],
      ),
    );
  }
}

class _ExpeditionStep extends StatelessWidget {
  const _ExpeditionStep({
    required this.index,
    required this.title,
    required this.description,
  });

  final String index;
  final String title;
  final String description;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Container(
          width: 48,
          height: 48,
          decoration: BoxDecoration(
            color: colorScheme.primary.withOpacity(0.12),
            borderRadius: BorderRadius.circular(14),
          ),
          child: Center(
            child: Text(
              index,
              style: TextStyle(
                color: colorScheme.primary,
                fontWeight: FontWeight.w700,
                letterSpacing: 1.2,
              ),
            ),
          ),
        ),
        const SizedBox(width: 12),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                title,
                style: Theme.of(context).textTheme.titleMedium?.copyWith(
                      fontWeight: FontWeight.w600,
                    ),
              ),
              const SizedBox(height: 4),
              Text(
                description,
                style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                      color: colorScheme.onSurface.withOpacity(0.7),
                    ),
              ),
            ],
          ),
        ),
      ],
    );
  }
}

class _TopoLinesPainter extends CustomPainter {
  _TopoLinesPainter(this.color);

  final Color color;

  @override
  void paint(Canvas canvas, Size size) {
    final linePaint = Paint()
      ..color = color
      ..style = PaintingStyle.stroke
      ..strokeWidth = 1;
    final starPaint = Paint()..color = color.withOpacity(0.45);
    final spacing = size.height / 10;

    for (var i = 0; i < 11; i++) {
      final path = Path();
      final baseY = spacing * i + (i.isEven ? 10 : -6);
      for (var j = 0; j <= 6; j++) {
        final x = size.width * (j / 6);
        final wave = math.sin((j / 6) * math.pi * 2 + i * 0.6) * 10;
        final y = baseY + wave;
        if (j == 0) {
          path.moveTo(x, y);
        } else {
          path.lineTo(x, y);
        }
      }
      canvas.drawPath(path, linePaint);
    }

    final rand = math.Random(12);
    for (var i = 0; i < 38; i++) {
      final dx = rand.nextDouble() * size.width;
      final dy = rand.nextDouble() * size.height;
      final radius = rand.nextDouble() * 1.4 + 0.4;
      canvas.drawCircle(Offset(dx, dy), radius, starPaint);
    }
  }

  @override
  bool shouldRepaint(covariant _TopoLinesPainter oldDelegate) {
    return oldDelegate.color != color;
  }
}
